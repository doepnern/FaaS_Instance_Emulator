package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandle(t *testing.T) {
	start := time.Now()

	payload := map[string]int{"n": 100, "size": 1000}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}
	var (
		w   = httptest.NewRecorder()
		req = httptest.NewRequest("POST", "http://localhost", bytes.NewReader(jsonPayload))
		res *http.Response
	)

	Handle(w, req)
	res = w.Result()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		t.Fatalf("unexpected response code: %v", res.StatusCode)
	}

	duration := time.Since(start)
	fmt.Printf("Response body: %s\n", body)
	fmt.Printf("TestHandle took %v\n", duration)
}
