package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockBrain is a mock implementation of the Brain for testing
// we implement just the methods needed by the Handler
type mockBrain struct{}

func (m *mockBrain) Reply(text string) (string, error) {
	return "instant mock reply for: " + text, nil
}

func (m *mockBrain) Learn(text string) error {
	return nil
}

func (m *mockBrain) Close() error {
	return nil
}

func TestPredict(t *testing.T) {
	// Create a mock brain
	mockBrain := &mockBrain{}

	// Create a handler with the mock brain
	handler := &Handler{
		Brain: mockBrain,
	}

	// Create a test request
	reqBody := RequestPayload{Text: "Test input"}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create a test server
	req := httptest.NewRequest(http.MethodPost, "/predict", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Call the handler
	handler.Predict(rec, req)

	// Check the response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Parse the response
	var resp ResponsePayload
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check the Reply method was used
	expectedReply := "instant mock reply for: Test input"
	if resp.Reply != expectedReply {
		t.Errorf("Expected reply %q, got %q", expectedReply, resp.Reply)
	}
}

func TestLearn(t *testing.T) {
	// Create a mock brain
	mockBrain := &mockBrain{}

	// Create a handler with the mock brain
	handler := &Handler{
		Brain: mockBrain,
	}

	// Create a test request
	reqBody := RequestPayload{Text: "Test input for learning"}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create a test server
	req := httptest.NewRequest(http.MethodPost, "/learn", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Call the handler
	handler.Learn(rec, req)

	// Check the response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}
