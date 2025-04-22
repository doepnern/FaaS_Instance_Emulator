package function

import (
	"encoding/json"
	factors "function/factorial"
	"net/http"
	"net/http/httputil"
)

type FactorialRequest struct {
	N *int `json:"n"`
}

type FactorialResponse struct {
	Data []int `json:"data"`
}

// Handle an HTTP Request.
func Handle(w http.ResponseWriter, r *http.Request) {
	/*
	 * YOUR CODE HERE
	 *
	 * Try running `go test`.  Add more test as you code in `handle_test.go`.
	 */

	_, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req FactorialRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.N == nil {
		http.Error(w, "Please pass number n as payload", http.StatusBadRequest)
		return
	}

	data := factors.Factors(*req.N)

	res := FactorialResponse{Data: data}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
