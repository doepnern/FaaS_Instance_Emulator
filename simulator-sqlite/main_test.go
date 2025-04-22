package main

import (
	"fmt"
	"testing"
	"time"
)

var cpuInstanceConstraint = Constraint{Resource: "cpu", Amount: 100, ResourcePerRequest: 100, AffectsUtilization: true}
var cpuGlobalConstraint = Constraint{Resource: "cpu", Amount: 800, ResourcePerRequest: 100, AffectsUtilization: true}
var rpsGlobalConstraint = Constraint{Resource: "rps", Amount: 999999, ResourcePerRequest: 1, AffectsUtilization: false}
var rpsInstanceConstraint = Constraint{Resource: "rps", Amount: 999999, ResourcePerRequest: 1, AffectsUtilization: false}

var globalConstraints = []Constraint{cpuGlobalConstraint, rpsGlobalConstraint}
var instanceConstraints = []Constraint{cpuInstanceConstraint, rpsInstanceConstraint}

var testFunctionProfile = FunctionProfile{
	Name: "test",
	Response: ResponseDistribution{
		Loc:   65.3328,
		Scale: 12.9785,
		S:     1.42729,
	},
	Replicas:            4,
	InstanceConstraints: instanceConstraints,
}

var testFunctionProfileFast = FunctionProfile{
	Name: "test",
	Response: ResponseDistribution{
		Loc:   2,
		Scale: 0,
		S:     0,
	},
	Replicas:            4,
	InstanceConstraints: instanceConstraints,
}

func TestNewSimulator(t *testing.T) {
	done := make(chan struct{})
	simulator := NewSimulator([]FunctionProfile{testFunctionProfile}, globalConstraints, 0)
	fmt.Println(simulator.db.GetAllRequests())
	go simulator.Start()
	simulator.AddRequest(Request{targetInstance: "test-0", done: done})
	simulator.AddRequest(Request{targetInstance: "test-0", done: done})
	simulator.AddRequest(Request{targetInstance: "test-0", done: done})
	simulator.AddRequest(Request{targetInstance: "test-0", done: done})
	simulator.AddRequest(Request{targetInstance: "test-0", done: done})
	simulator.AddRequest(Request{targetInstance: "test-1", done: done})

	requests := simulator.GetAllRequests()
	for _, req := range requests {
		fmt.Println(req)
	}
	fmt.Println("waiting for done")
	simulator.WaitAllDone()
	fmt.Println("done")
	simulator.Stop()
	fmt.Println(simulator)
}

func TestManyRequests(t *testing.T) {
	done := make(chan struct{})
	globalStore := NewSimulator([]FunctionProfile{testFunctionProfileFast}, globalConstraints, 0)
	go globalStore.Start()
	var avgAddRequestTime float64
	var count int = 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 5000; j++ {
			start := time.Now()
			globalStore.AddRequest(Request{targetInstance: "test-" + fmt.Sprint(i)})
			duration := time.Since(start)
			avgAddRequestTime += float64(duration)
			count++
			if duration > 3*time.Millisecond {
				t.Errorf("AddRequest took too long: %v", duration)
			}

			// fmt.Printf("Added request in %.3f ms\n", float64(duration)/float64(time.Millisecond))
			time.Sleep(1 * time.Millisecond)
		}
	}
	globalStore.AddRequest(Request{targetInstance: "test-1", done: done})

	fmt.Println(globalStore)
	fmt.Println("Avg AddRequest time(ms): ", avgAddRequestTime/float64(count)/float64(time.Millisecond))
}

func TestAdd(t *testing.T) {
	loc := 5.3328
	scale := 12.9785
	s := 1.42729
	res := LognormalGenerator(loc, scale, s)
	fmt.Println(res)
}

var globalHostConsraint2 = []Constraint{{Resource: "cpu", Amount: 800, ResourcePerRequest: 100, AffectsUtilization: true}}
var instanceConstraints2 = []Constraint{{Resource: "cpu", Amount: 800, ResourcePerRequest: 100, AffectsUtilization: true}}

var testFunctionProfileHostConstrain = FunctionProfile{
	Name: "test",
	Response: ResponseDistribution{
		Loc:   10.3328,
		Scale: 12.9785,
		S:     1.42729,
	},
	Replicas:            4,
	InstanceConstraints: instanceConstraints2,
}

func TestHostConstraint(t *testing.T) {
	simulator := NewSimulator([]FunctionProfile{testFunctionProfileHostConstrain}, globalHostConsraint2, 0)
	go simulator.Start()
	for i := 0; i < 3; i++ {
		for j := 0; j < 50; j++ {
			simulator.AddRequest(Request{targetInstance: "test-" + fmt.Sprint(i), done: make(chan struct{})})
			time.Sleep(1 * time.Millisecond)
		}
	}
	simulator.WaitAllDone()
	simulator.Stop()

}
