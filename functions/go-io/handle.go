package function

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
)

type IoRequest struct {
	N    *int `json:"n"`
	Size *int `json:"size"`
}

type IoResponse struct {
	Data string `json:"data"`
}

// Handle an HTTP Request.
func Handle(w http.ResponseWriter, r *http.Request) {
	_, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req IoRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.N == nil {
		http.Error(w, "Please pass number n as payload", http.StatusBadRequest)
		return
	}
	if req.Size == nil {
		http.Error(w, "Please pass size as payload", http.StatusBadRequest)
		return
	}
	n := *req.N
	size := *req.Size

	// Reverse the string
	totalChars := readWrite(n, size)

	res := IoResponse{Data: strconv.Itoa(totalChars)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Writes n files with len size and then reads every file. Hinspired by https://github.com/faas-benchmarking/faasdom/blob/master/aws/src/go/go_filesystem/filesystem.go
func readWrite(n int, size int) int {
	rnd := rand.Intn(90000000) + 1000000

	if _, err := os.Stat("/tmp/test"); os.IsNotExist(err) {
		os.Mkdir("/tmp/test", 0777)
	}

	if _, err := os.Stat("/tmp/test/" + strconv.Itoa(rnd)); os.IsNotExist(err) {
		os.Mkdir("/tmp/test/"+strconv.Itoa(rnd), 0777)
	}

	var text string = ""

	for i := 0; i < size; i++ {
		text += "A"
	}
	baseFile := "/tmp/test/" + strconv.Itoa(rnd) + "/0.txt"

	f, err := os.Create(baseFile)
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString(text)
	f.Close()

	for i := 1; i < n; i++ {
		readFile, err := os.Open(baseFile)
		if err != nil {
			log.Fatal(err)
		}
		copyFile, err := os.Create("/tmp/test/" + strconv.Itoa(rnd) + "/" + strconv.Itoa(i) + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		_, err = readFile.WriteTo(copyFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	// delete folder
	err = os.RemoveAll("/tmp/test/" + strconv.Itoa(rnd))
	if err != nil {
		log.Fatal(err)
	}
	return n * size
}
