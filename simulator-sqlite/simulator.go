package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/doepnern/faas-simulator/db"
	"github.com/doepnern/faas-simulator/logs"
)

type GlobalStore struct {
	resources      map[string]float64
	instances      map[string]*InstanceStore
	constraints    []Constraint
	stopped        bool
	networkLatency float64
	mu             *sync.Mutex
	db             *db.Database
}

func (gs GlobalStore) String() string {
	result := "GlobalStore:\n"
	result += "  Resources:\n"
	for k, v := range gs.resources {
		result += fmt.Sprintf("    %s: %.2f\n", k, v)
	}
	result += "  Instances:\n"
	for k, v := range gs.instances {
		result += fmt.Sprintf("    %s:\n", k)
		for rk, rv := range v.resources {
			result += fmt.Sprintf("      %s: %.2f\n", rk, rv)
		}
	}
	return result
}

type Request struct {
	targetInstance string
	done           chan struct{}
}

type InstanceStore struct {
	resources            map[string]float64
	responseDistribution ResponseDistribution
	constraints          []Constraint
	requestChannels      map[int]*chan struct{}
	mu                   sync.Mutex
	responsePayload      interface{}
}

func NewSimulator(functionProfiles []FunctionProfile, globalConstraints []Constraint, networkLatency float64) *GlobalStore {
	db, err := db.New()
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic(errors.New("database not initialized"))
	}
	if db.Connection == nil {
		panic(errors.New("database connection not initialized"))
	}

	globalStore := GlobalStore{
		resources:      make(map[string]float64),
		instances:      make(map[string]*InstanceStore),
		constraints:    globalConstraints,
		stopped:        false,
		networkLatency: networkLatency,
		mu:             &sync.Mutex{},
		db:             db,
	}

	for _, functionProfile := range functionProfiles {
		numInstances := functionProfile.Replicas
		for i := 0; i < numInstances; i++ {
			instance := InstanceStore{resources: make(map[string]float64), responseDistribution: functionProfile.Response, constraints: functionProfile.InstanceConstraints, responsePayload: functionProfile.ResponsePayload, requestChannels: make(map[int]*chan struct{})}
			for _, constraint := range functionProfile.InstanceConstraints {
				instance.resources[constraint.Resource] = constraint.Amount
			}
			globalStore.instances[fmt.Sprintf("%s-%d", functionProfile.Name, i)] = &instance
		}
		for _, constraint := range globalStore.constraints {
			globalStore.resources[constraint.Resource] = constraint.Amount
		}
	}

	return &globalStore
}

func GetAverageResourceUsageForConstraints(constraints []Constraint, resourceUsage map[string]float64) float64 {
	var total float64
	var count int
	for _, constraint := range constraints {
		if !constraint.AffectsUtilization {
			continue
		}
		averagePercentOfMax := resourceUsage[constraint.Resource] / constraint.Amount
		total += averagePercentOfMax
		count++
	}
	return total / float64(count)
}

func (globalStore *GlobalStore) HandleWaitingRequests() {
	for !globalStore.stopped {
		instanceKeys, err := globalStore.db.GetInstanceOrder()
		if err != nil {
			panic(err)
		}
		for _, name := range instanceKeys {
			reqs, err := globalStore.db.GetUnhandledRequestsByInstance(name)
			if err != nil {
				panic(err)
			}
			if len(reqs) == 0 {
				continue
			}
			req := reqs[0]
			canHandle, reason := globalStore.InstanceCanHandleRequest(req)
			if canHandle {
				now := time.Now().UnixNano()
				// resourceUsgageLastSecond, err := globalStore.db.GetAverageResourceUsageInTimeFrame(req.TargetInstance, now-1e9, now)
				// if err != nil {
				// 	panic(err)
				// }

				// averageUsageInstance := GetAverageResourceUsageForConstraints(globalStore.instances[req.TargetInstance].constraints, resourceUsgageLastSecond)
				// averageUsageHost := GetAverageResourceUsageForConstraints(globalStore.constraints, resourceUsgageLastSecond)

				// logs.DefaultLogger.Info(fmt.Sprintf("Average usage for instance %s: %f", req.TargetInstance, averageUsageInstance))
				// logs.DefaultLogger.Info(fmt.Sprintf("Average usage for host: %f", averageUsageHost))
				globalStore.db.RemoveStaleResourceUsage(now)

				instance := globalStore.instances[req.TargetInstance]
				responseTime := LognormalGenerator(instance.responseDistribution.Loc, instance.responseDistribution.Scale, instance.responseDistribution.S) - globalStore.networkLatency
				logs.DefaultLogger.Info(fmt.Sprintf("Planned Response time for request %d: %f", req.ID, responseTime))
				globalStore.db.UpdateRequest(req.ID, time.Now().Add(time.Duration(responseTime*1e6)).UnixNano())

				for _, constraint := range instance.constraints {
					durationMultiplier := 1.0
					if constraint.DurationMultiplicator != nil {
						durationMultiplier = *constraint.DurationMultiplicator
					}
					now := time.Now().UnixNano()

					globalStore.db.AddResourceUsage(constraint.Resource, constraint.ResourcePerRequest, now, int64(float64(time.Now().Add(time.Duration(responseTime*1e6*durationMultiplier)).UnixNano())), req.TargetInstance)
					globalStore.db.AddResourceUsage(constraint.Resource, constraint.ResourcePerRequest, now, int64(float64(time.Now().Add(time.Duration(responseTime*1e6*durationMultiplier)).UnixNano())), "host")
				}
			} else if reason == "host" {
				break
			}
		}
		time.Sleep(100 * time.Nanosecond)
	}
}

func (globalStore *GlobalStore) HandleResponses() {
	for !globalStore.stopped {
		currentTime := time.Now().UnixNano()

		reqs, err := globalStore.db.GetHandledRequests(currentTime)
		if err != nil {
			panic(err)
		}
		_, err = globalStore.db.RemoveHandledRequests(currentTime)
		if err != nil {
			panic(err)
		}

		for _, req := range reqs {
			instance := globalStore.instances[req.TargetInstance]
			done := *instance.requestChannels[req.ID]
			if done == nil {
				continue
			}
			close(done)
			instance.mu.Lock()
			delete(instance.requestChannels, req.ID)
			instance.mu.Unlock()
		}
		time.Sleep(10 * time.Nanosecond)
	}
}

func (globalStore *GlobalStore) InstanceCanHandleRequest(req db.Request) (bool, string) {
	instance := globalStore.instances[req.TargetInstance]
	if instance == nil {
		panic("Instance not found")
	}
	currentTime := time.Now().UnixNano()

	resourceUsage, err := globalStore.db.GetResourceUsageByResourceForTarget(req.TargetInstance, currentTime)
	if err != nil {
		panic(err)
	}
	resourceUsageHost, err := globalStore.db.GetResourceUsageByResourceForTarget("host", currentTime)
	if err != nil {
		panic(err)
	}
	instance.mu.Lock()
	defer instance.mu.Unlock()
	// check if host can handle request
	for _, constraint := range globalStore.constraints {
		if constraint.Amount-resourceUsageHost[constraint.Resource] < constraint.ResourcePerRequest {
			logs.DefaultLogger.Debug(fmt.Sprintln("Resource usage for host", req.TargetInstance, "and resource", constraint.Resource, "is", resourceUsageHost[constraint.Resource], "but needs to be lower than", constraint.Amount-constraint.ResourcePerRequest))
			return false, "host"
		}
	}
	// check if instance can handle request
	for _, constraint := range instance.constraints {
		if constraint.Amount-resourceUsage[constraint.Resource] < constraint.ResourcePerRequest {
			logs.DefaultLogger.Debug(fmt.Sprintln("Resource usage for instance", req.TargetInstance, "and resource", constraint.Resource, "is", resourceUsage[constraint.Resource], "but needs to be lower than", constraint.Amount-constraint.ResourcePerRequest))
			return false, "instance"
		}
	}

	return true, ""
}

func (globalStore *GlobalStore) Start() {
	go globalStore.HandleWaitingRequests()
	go globalStore.HandleResponses()
}

func (globalStore *GlobalStore) WaitAllDone() {
	for _, instance := range globalStore.instances {
		for _, done := range instance.requestChannels {
			d := *done
			if d == nil {
				continue
			}
			_, ok := <-d
			if !ok {
				continue
			}
		}
	}
}

func (globalStore *GlobalStore) Stop() {
	globalStore.stopped = true
	time.Sleep(100 * time.Millisecond)
	globalStore.db.Close()
}

func (globalStore *GlobalStore) AddRequest(request Request) int {
	id, err := globalStore.db.AddRequest(request.targetInstance)
	if err != nil {
		panic(err)
	}
	instance := globalStore.instances[request.targetInstance]
	if instance == nil {
		panic("Instance not found")
	}
	instance.mu.Lock()
	defer instance.mu.Unlock()
	instance.requestChannels[id] = &request.done
	return id
}

func (globalStore *GlobalStore) GetAllRequests() []db.Request {
	requests, err := globalStore.db.GetAllRequests()
	if err != nil {
		panic(err)
	}
	return requests
}
