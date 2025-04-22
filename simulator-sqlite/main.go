package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/doepnern/faas-simulator/logs"
)

func main() {
	definition, err := ReadSimulatorDefinition("./definition.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	simulator := NewSimulator(definition.FunctionProfiles, definition.GlobalConstraints, definition.NetworkLatency)
	fmt.Println("Starting simulator")
	fmt.Println(simulator)
	go simulator.Start()
	m := http.NewServeMux()
	srv := http.Server{Addr: ":8080", Handler: m}
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//parse instance name
		start := time.Now()
		instanceName := r.URL.Query().Get("name")
		done := make(chan struct{})
		rq := Request{targetInstance: instanceName, done: done}
		id := simulator.AddRequest(rq)
		// wait for done channel to be closed
		<-done
		end := time.Now()
		logs.DefaultLogger.Info(fmt.Sprintf("Request %d took: %v", id, end.Sub(start)))

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(simulator.instances[instanceName].responsePayload)
	})
	srv.ListenAndServe()
}
