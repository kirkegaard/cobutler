package api

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
)

// Brain defines the interface required by the API handlers
type Brain interface {
	Reply(text string) (string, error)
	Learn(text string) error
	RememberCompletion(context, completion string)
	EnableCache()
	DisableCache()
	Close() error
}

// RequestPayload represents the incoming JSON request
type RequestPayload struct {
	Text      string  `json:"text"`
	MaxWords  int     `json:"max_words,omitempty"`
	Precision float64 `json:"precision,omitempty"`
	Context   string  `json:"context,omitempty"`
	UseCache  bool    `json:"use_cache,omitempty"`
}

// ResponsePayload represents the outgoing JSON response
type ResponsePayload struct {
	Reply string `json:"reply"`
}

// Handler contains the HTTP handlers for the API
type Handler struct {
	Brain Brain
}

// NewHandler creates a new Handler
func NewHandler(brain Brain) *Handler {
	return &Handler{
		Brain: brain,
	}
}

// SetupRoutes configures the HTTP routes for the application
func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/predict", h.Predict)
	mux.HandleFunc("/learn", h.Learn)
}

// Predict handles requests to generate predictions from the brain
func (h *Handler) Predict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		slog.Warn("Method not allowed", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Invalid request", "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	slog.Info("Received predict request",
		"text_length", len(req.Text),
		"max_words", req.MaxWords,
		"precision", req.Precision,
		"use_cache", req.UseCache)

	// Check if cache setting should be modified
	if brainWithCache, ok := h.Brain.(interface {
		EnableCache()
		DisableCache()
	}); ok {
		if req.UseCache {
			slog.Info("Enabling cache for request")
			brainWithCache.EnableCache()
		} else {
			slog.Info("Disabling cache for request")
			brainWithCache.DisableCache()
		}
	}

	// Extract code-specific information
	filetype, processedText := extractCodeMetadata(req.Text)

	// Get multiple replies and select based on precision rating
	var reply string
	var err error

	// Default precision if not specified
	precision := req.Precision
	if precision <= 0 {
		precision = 0.7 // Default precision
	}
	if precision > 1 {
		precision = 1.0 // Cap at 1.0
	}

	// For high precision, generate multiple responses and find most common elements
	if precision > 0.7 {
		// Generate multiple responses - more for higher precision
		numResponses := 3
		if precision > 0.9 {
			numResponses = 5
		}

		replies := make([]string, numResponses)
		for i := 0; i < numResponses; i++ {
			replies[i], err = h.Brain.Reply(processedText)
			if err != nil {
				slog.Error("Failed to generate reply", "error", err)
				http.Error(w, "Failed to generate reply", http.StatusInternalServerError)
				return
			}
		}

		// Select reply based on precision
		reply = selectReplyByPrecision(replies, precision)
	} else {
		// For lower precision, just get a single response (more creative)
		reply, err = h.Brain.Reply(processedText)
		if err != nil {
			slog.Error("Failed to generate reply", "error", err)
			http.Error(w, "Failed to generate reply", http.StatusInternalServerError)
			return
		}
	}

	// Post-process the reply based on filetype and improve code completion
	reply = postProcessCodeReply(reply, filetype)

	// Apply max_words limit if provided
	if req.MaxWords > 0 {
		reply = limitWords(reply, req.MaxWords)
	}

	resp := ResponsePayload{Reply: reply}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	slog.Info("Predict request succeeded", "response_length", len(reply))
}

// Learn handles requests to train the brain with new text
func (h *Handler) Learn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		slog.Warn("Method not allowed", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Invalid request", "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	slog.Info("Received learn request", "text_length", len(req.Text))

	// Process the text, removing any special markers
	_, cleanText := extractCodeMetadata(req.Text)

	if err := h.Brain.Learn(cleanText); err != nil {
		slog.Error("Failed to learn", "error", err)
		http.Error(w, "Failed to learn", http.StatusInternalServerError)
		return
	}

	// If there's a lastContext and this is a response to it, remember this completion
	if len(req.Context) > 0 && len(cleanText) > 0 {
		h.Brain.RememberCompletion(req.Context, cleanText)
		slog.Info("Remembered completion for context", "context_length", len(req.Context))
	}

	w.WriteHeader(http.StatusOK)
	slog.Info("Learn request succeeded")
}

// limitWords restricts a string to a maximum number of words
func limitWords(text string, maxWords int) string {
	if maxWords <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) <= maxWords {
		return text
	}

	return strings.Join(words[:maxWords], " ")
}

// selectReplyByPrecision picks a response based on precision level
func selectReplyByPrecision(replies []string, precision float64) string {
	if len(replies) == 0 {
		return ""
	}

	if len(replies) == 1 {
		return replies[0]
	}

	// Higher precision values choose shorter, more conservative replies
	if precision > 0.9 {
		// Find the shortest reply
		shortestIdx := 0
		shortestLen := len(strings.Fields(replies[0]))

		for i := 1; i < len(replies); i++ {
			currLen := len(strings.Fields(replies[i]))
			if currLen < shortestLen {
				shortestLen = currLen
				shortestIdx = i
			}
		}

		return replies[shortestIdx]
	} else if precision > 0.7 {
		// Find a reply that's not too long or too short (middle of the pack)
		lengths := make([]int, len(replies))
		for i, reply := range replies {
			lengths[i] = len(strings.Fields(reply))
		}

		// Sort the indices by length
		type indexedLength struct {
			idx    int
			length int
		}

		sortedLengths := make([]indexedLength, len(lengths))
		for i, length := range lengths {
			sortedLengths[i] = indexedLength{idx: i, length: length}
		}

		// Sort by length
		for i := 0; i < len(sortedLengths); i++ {
			for j := i + 1; j < len(sortedLengths); j++ {
				if sortedLengths[i].length > sortedLengths[j].length {
					sortedLengths[i], sortedLengths[j] = sortedLengths[j], sortedLengths[i]
				}
			}
		}

		// Pick the middle one
		midIdx := sortedLengths[len(sortedLengths)/2].idx
		return replies[midIdx]
	} else {
		// Lower precision, more randomness - pick randomly from all responses
		return replies[rand.Intn(len(replies))]
	}
}

// extractCodeMetadata extracts filetype and other code metadata from the text
func extractCodeMetadata(text string) (string, string) {
	// Default filetype
	filetype := "text"

	// Check for filetype marker in the text
	filetypeRegex := regexp.MustCompile(`(?m)^// FILETYPE: (\w+)$`)
	matches := filetypeRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		filetype = matches[1]
		// Remove the filetype marker from the text
		text = filetypeRegex.ReplaceAllString(text, "")
	}

	// Clean out any other special markers we've added
	text = regexp.MustCompile(`(?m)^// AFTER CURSOR:.*$`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`(?m)^// CONTEXT AFTER:.*$`).ReplaceAllString(text, "")

	// Clean up multiple blank lines
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")

	return filetype, strings.TrimSpace(text)
}

// postProcessCodeReply improves the code quality of replies
func postProcessCodeReply(reply, filetype string) string {
	// Skip processing for non-code filetypes
	if filetype == "text" || filetype == "" {
		return reply
	}

	// Fix common code formatting issues based on filetype
	switch filetype {
	case "go":
		// Fix unmatched brackets or parentheses
		reply = fixBracketBalance(reply, "{", "}")
		reply = fixBracketBalance(reply, "(", ")")

		// Ensure proper spacing after keywords
		reply = regexp.MustCompile(`(if|for|switch|func)\(`).ReplaceAllString(reply, "$1 (")

	case "javascript", "typescript", "jsx", "tsx":
		// Fix unmatched brackets, parentheses, or template literals
		reply = fixBracketBalance(reply, "{", "}")
		reply = fixBracketBalance(reply, "(", ")")
		reply = fixBracketBalance(reply, "[", "]")

		// Fix arrow functions
		reply = regexp.MustCompile(`(\w+)\s*=>\s*{([^}]*)$`).ReplaceAllString(reply, "$1 => {$2}")

	case "python":
		// Fix indentation issues
		lines := strings.Split(reply, "\n")
		if len(lines) > 1 {
			// Check if we need to adjust indentation
			if strings.HasPrefix(lines[0], "    ") || strings.HasPrefix(lines[0], "\t") {
				// The reply starts indented, which might be incorrect
				indent := ""
				for _, c := range lines[0] {
					if c == ' ' || c == '\t' {
						indent += string(c)
					} else {
						break
					}
				}
				// Remove the indentation from all lines
				for i := range lines {
					if strings.HasPrefix(lines[i], indent) {
						lines[i] = strings.TrimPrefix(lines[i], indent)
					}
				}
				reply = strings.Join(lines, "\n")
			}
		}

	case "lua":
		// Fix function declarations
		reply = regexp.MustCompile(`function\s*([a-zA-Z0-9_.]+)\s*\(`).ReplaceAllString(reply, "function $1(")

		// Fix 'end' keyword if missing
		if strings.Contains(reply, "function ") && !strings.Contains(reply, "end") {
			reply += "\nend"
		}
	}

	return reply
}

// fixBracketBalance ensures balanced brackets/parentheses
func fixBracketBalance(text, opening, closing string) string {
	openCount := strings.Count(text, opening)
	closeCount := strings.Count(text, closing)

	// Add missing closing brackets
	if openCount > closeCount {
		text += strings.Repeat(closing, openCount-closeCount)
	}

	return text
}
