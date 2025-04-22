package function

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"strconv"
)

type StringReversalRequest struct {
	N *int `json:"n"`
}

type StringReversalResponse struct {
	Data string `json:"data"`
}

// Handle an HTTP Request.
func Handle(w http.ResponseWriter, r *http.Request) {
	_, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req StringReversalRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.N == nil {
		http.Error(w, "Please pass number n as payload", http.StatusBadRequest)
		return
	}
	length := *req.N

	// Create the string "12345678..." with length n
	payload := ""
	for i := 1; i <= length; i++ {
		payload += strconv.Itoa(i)
	}

	// Reverse the string
	data := reverseString(payload)

	res := StringReversalResponse{Data: data}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Helper function to reverse a string
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
