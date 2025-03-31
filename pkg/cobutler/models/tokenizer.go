package models

import (
	"regexp"
	"strings"
)

// Tokenizer represents an interface for different tokenization strategies
type Tokenizer interface {
	Split(text string) []string
}

// CobeTokenizer implements the Cobe tokenization strategy
type CobeTokenizer struct {
	sentenceSplitter *regexp.Regexp
	punctuation      *regexp.Regexp
}

// NewCobeTokenizer creates a new CobeTokenizer
func NewCobeTokenizer() *CobeTokenizer {
	return &CobeTokenizer{
		sentenceSplitter: regexp.MustCompile(`[.!?]+\s+`),
		punctuation:      regexp.MustCompile(`[,.!?;:]`),
	}
}

// Split splits the text into tokens
func (t *CobeTokenizer) Split(text string) []string {
	// Normalize whitespace
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	var tokens []string
	var buffer []rune

	for _, char := range text {
		if char == ' ' {
			if len(buffer) > 0 {
				tokens = append(tokens, string(buffer))
				buffer = nil
			}
			tokens = append(tokens, " ")
		} else if t.isPunctuation(char) {
			if len(buffer) > 0 {
				tokens = append(tokens, string(buffer))
				buffer = nil
			}
			tokens = append(tokens, string(char))
		} else {
			buffer = append(buffer, char)
		}
	}

	if len(buffer) > 0 {
		tokens = append(tokens, string(buffer))
	}

	return tokens
}

// isPunctuation checks if a character is considered punctuation
func (t *CobeTokenizer) isPunctuation(char rune) bool {
	return strings.ContainsRune(",.!?;:", char)
}

// MegaHALTokenizer implements the MegaHAL tokenization strategy
type MegaHALTokenizer struct{}

// NewMegaHALTokenizer creates a new MegaHALTokenizer
func NewMegaHALTokenizer() *MegaHALTokenizer {
	return &MegaHALTokenizer{}
}

// Split splits the text into tokens
func (t *MegaHALTokenizer) Split(text string) []string {
	var tokens []string
	var buffer []rune

	inWord := false
	for _, char := range text {
		isAlphaNum := isAlphanumeric(char)

		if isAlphaNum != inWord {
			if len(buffer) > 0 {
				tokens = append(tokens, string(buffer))
				buffer = nil
			}
			inWord = isAlphaNum
		}

		if char == ' ' {
			if len(buffer) > 0 {
				tokens = append(tokens, string(buffer))
				buffer = nil
			}
			tokens = append(tokens, " ")
			inWord = false
		} else {
			buffer = append(buffer, char)
		}
	}

	if len(buffer) > 0 {
		tokens = append(tokens, string(buffer))
	}

	return tokens
}

// isAlphanumeric checks if a character is a letter or digit
func isAlphanumeric(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}
