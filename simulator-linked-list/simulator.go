package main

import (
	"container/list"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Request struct {
	targetInstance string
	done           *chan struct{}
	receivedAt     time.Time
	processedAt    *time.Time
	id             string
	isClosed       bool
}

type Simulator struct {
	instances      map[string]*InstanceStore
	stopped        bool
	networkLatency float64
	requestQueue   list.List
	rqMu           *sync.Mutex
	processedQueue list.List
	prqMu          *sync.Mutex
}

func (sim Simulator) GetHostRps() int {
	sum := 0
	sim.prqMu.Lock()
	for e := sim.processedQueue.Front(); e != nil; e = e.Next() {
		request := e.Value.(Request)
		if request.processedAt != nil && request.processedAt.After(time.Now().Add(-1*time.Second)) {
			sum++
		}
	}
	sim.prqMu.Unlock()
	return sum
}

func (sim Simulator) GetInstanceRps(instanceName string) int {
	sum := 0
	sim.prqMu.Lock()
	for e := sim.processedQueue.Front(); e != nil; e = e.Next() {
		request := e.Value.(Request)
		if request.targetInstance == instanceName && request.processedAt != nil && request.processedAt.After(time.Now().Add(-1*time.Second)) {
			sum++
		}
	}
	sim.prqMu.Unlock()
	return sum
}

func (gs Simulator) String() string {
	result := "GlobalStore:\n"
	result += "  Resources:\n"
	result += "  Instances:\n"
	for k := range gs.instances {
		result += fmt.Sprintf("    %s:\n", k)
	}
	return result
}

func NewSimulator(functionProfiles []FunctionProfile, networkLatency float64) *Simulator {
	simulatorState := Simulator{
		instances:      make(map[string]*InstanceStore),
		stopped:        false,
		networkLatency: networkLatency,
		requestQueue:   list.List{},
		rqMu:           &sync.Mutex{},
		processedQueue: list.List{},
		prqMu:          &sync.Mutex{},
	}

	for _, functionProfile := range functionProfiles {
		numInstances := functionProfile.Replicas
		for i := 0; i < numInstances; i++ {
			instance := InstanceStore{responseDistributions: functionProfile.Response, responsePayload: functionProfile.ResponsePayload, maxRps: int(functionProfile.MaxRps), maxHostRps: int(functionProfile.MaxHostRps)}
			simulatorState.instances[fmt.Sprintf("%s-%d", functionProfile.Name, i)] = &instance
		}
	}

	return &simulatorState
}

func (sim *Simulator) HandleWaitingRequests() {
	for !sim.stopped {
		blockedInstance := make(map[string]bool)
		time.Sleep(1000 * time.Nanosecond)
		sim.rqMu.Lock()
		for e := sim.requestQueue.Front(); e != nil; e = e.Next() {
			if e.Value == nil {
				continue
			}
			request := e.Value.(Request)
			if blockedInstance[request.targetInstance] {
				continue
			}
			instance := sim.GetInstance(request.targetInstance)
			hostRps := sim.GetHostRps()
			instanceRps := sim.GetInstanceRps(request.targetInstance)
			canHandle, reason := instance.InstanceCanHandleRequest(request, hostRps, instanceRps)
			if canHandle {
				loadInstance := float64(instanceRps) / float64(instance.maxRps)
				loadHost := float64(hostRps) / float64(instance.maxHostRps)
				load := math.Max(loadInstance, loadHost)
				distributionForLoad := int(math.Min(math.Floor(load*float64((len(instance.responseDistributions)))), float64(len(instance.responseDistributions)-1)))
				// fmt.Println("Load level: ", distributionForLoad)
				responseDistribution := instance.responseDistributions[distributionForLoad]
				responseTime := LognormalGenerator(responseDistribution.Loc, responseDistribution.Scale, responseDistribution.S) - sim.networkLatency
				// fmt.Println("Response time for request: ", request.id, " is: ", responseTime)
				processedAt := time.Now().Add(time.Duration(responseTime) * time.Millisecond)
				request.processedAt = &processedAt
				sim.prqMu.Lock()
				sim.processedQueue.PushBack(request)
				sim.prqMu.Unlock()
				sim.requestQueue.Remove(e)
				break
			}
			if reason == "host" {
				break
			}
			if reason == "instance" {
				blockedInstance[request.targetInstance] = true
			}
		}
		sim.rqMu.Unlock()
	}
}

func (sim *Simulator) HandleProcessedRequests() {
	for !sim.stopped {
		time.Sleep(100 * time.Nanosecond)
		sim.prqMu.Lock()
		for e := sim.processedQueue.Front(); e != nil; e = e.Next() {
			request := e.Value.(Request)
			if request.processedAt == nil {
				panic("Unprocessed request in processed queue")
			}

			if request.processedAt.Before(time.Now()) {
				if !request.isClosed {
					request.isClosed = true
					e.Value = request
					if request.done != nil {
						close(*request.done)
					}
				}
				// keep processed requests for 1 second
				if request.isClosed && request.processedAt.Before(time.Now().Add(-1*time.Second)) {
					sim.processedQueue.Remove(e)
				}
			}
		}
		sim.prqMu.Unlock()
	}
}

func (globalStore *Simulator) Start() {
	go globalStore.HandleWaitingRequests()
	go globalStore.HandleProcessedRequests()
}

func (globalStore *Simulator) Stop() {
	globalStore.stopped = true
}

func (globalStore *Simulator) AddRequest(request Request) (string, error) {
	if request.id == "" {
		request.id = uuid.New().String()
	}
	if request.receivedAt.IsZero() {
		request.receivedAt = time.Now()
	}
	globalStore.rqMu.Lock()
	defer globalStore.rqMu.Unlock()
	if globalStore.requestQueue.Len() > 1000 {
		return "", ErrorInstanceFull{}
	}
	globalStore.requestQueue.PushBack(request)
	return request.id, nil
}

type ErrorInstanceFull struct {
}

func (e ErrorInstanceFull) Error() string {
	return "Instance is full"
}
