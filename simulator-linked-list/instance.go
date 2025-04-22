package main

import "fmt"

type InstanceStore struct {
	responseDistributions []ResponseDistribution
	responsePayload       interface{}
	maxRps                int
	maxHostRps            int
}

func (sim Simulator) GetInstance(instanceName string) *InstanceStore {
	instance := sim.instances[instanceName]
	if instance == nil {
		fmt.Println(sim.instances)
		panic(fmt.Sprintf("Instance %s not found", instanceName))
	}
	return instance
}

func (instance *InstanceStore) InstanceCanHandleRequest(rq Request, hostRps int, instanceRps int) (bool, string) {
	fmt.Println(instanceRps, hostRps)
	if hostRps >= instance.maxHostRps {
		return false, "host"
	}
	if instanceRps >= instance.maxRps {
		return false, "instance"
	}
	return true, ""
}
