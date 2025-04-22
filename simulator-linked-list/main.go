package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Please provide the name of the definition file")
		os.Exit(1)
	}
	definitionFile := args[0]
	slog.Debug(fmt.Sprintf("Reading definition file: %s", definitionFile))
	definition, err := ReadSimulatorDefinition(fmt.Sprintf("./%s.json", definitionFile))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	simulator := NewSimulator(definition.FunctionProfiles, definition.NetworkLatency)
	fmt.Println("Starting simulator")
	fmt.Println(simulator)
	go simulator.Start()
	m := http.NewServeMux()
	srv := http.Server{Addr: ":8089", Handler: m}
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//parse instance name
		start := time.Now()
		instanceName := r.URL.Query().Get("name")
		done := make(chan struct{})
		rq := Request{targetInstance: instanceName, done: &done, receivedAt: time.Now(), id: uuid.New().String()}
		_, err := simulator.AddRequest(rq)
		if err != nil {
			slog.Error(fmt.Sprintf("Error adding request: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// wait for done channel to be closed
		<-done
		end := time.Now()
		slog.Debug(fmt.Sprintf("Request %s took: %v", rq.id, end.Sub(start)))
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(simulator.instances[instanceName].responsePayload)
	})
	srv.ListenAndServe()
}
