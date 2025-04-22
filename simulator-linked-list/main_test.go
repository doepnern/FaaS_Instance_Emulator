package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

var testFunctionProfile = FunctionProfile{
	Name: "test",
	Response: []ResponseDistribution{
		{Loc: 1.3328,
			Scale: 12.9785,
			S:     1.42729},
	},
	Replicas:   4,
	MaxRps:     100,
	MaxHostRps: 1000,
}

func TestNewSimulator(t *testing.T) {
	done := make(chan struct{})
	globalStore := NewSimulator([]FunctionProfile{testFunctionProfile}, 0)
	go globalStore.Start()
	globalStore.AddRequest(Request{targetInstance: "test-0"})
	globalStore.AddRequest(Request{targetInstance: "test-0"})
	globalStore.AddRequest(Request{targetInstance: "test-0"})
	globalStore.AddRequest(Request{targetInstance: "test-0"})
	globalStore.AddRequest(Request{targetInstance: "test-0"})
	globalStore.AddRequest(Request{targetInstance: "test-1", done: &done})

	<-done
	fmt.Println(globalStore)
	globalStore.Stop()
}

func TestManyRequests(t *testing.T) {
	done := make(chan struct{})
	globalStore := NewSimulator([]FunctionProfile{testFunctionProfile}, 0)
	go globalStore.Start()
	for i := 0; i < 3; i++ {
		for j := 0; j < 1000; j++ {
			globalStore.AddRequest(Request{targetInstance: "test-" + fmt.Sprint(i), receivedAt: time.Now(), id: uuid.New().String()})
			time.Sleep(100 * time.Nanosecond)
		}

	}
	globalStore.AddRequest(Request{targetInstance: "test-1", done: &done})
	fmt.Println("Waiting for done")
	<-done
	fmt.Println(globalStore)
	globalStore.Stop()
	fmt.Println("Stopped")

}

func TestAdd(t *testing.T) {
	loc := 5.3328
	scale := 12.9785
	s := 1.42729
	res := LognormalGenerator(loc, scale, s)
	fmt.Println(res)

}
