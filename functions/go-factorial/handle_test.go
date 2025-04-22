package function

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandle ensures that Handle executes without error and returns the
// HTTP 200 status code indicating no errors.
func TestHandle(t *testing.T) {
	payload := map[string]int{"n": 5}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}
	var (
		w   = httptest.NewRecorder()
		req = httptest.NewRequest("POST", "http://localhost",  bytes.NewReader(jsonPayload))
		res *http.Response
	)

	Handle(w, req)
	res = w.Result()
	defer res.Body.Close()

	if res.StatusCode != 200 {
		t.Fatalf("unexpected response code: %v", res.StatusCode)
	}
}
