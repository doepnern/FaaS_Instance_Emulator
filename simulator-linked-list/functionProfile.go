package main

import (
	"encoding/json"
	"os"
)

type ResponseDistribution struct {
	Loc   float64 `json:"loc"`
	Scale float64 `json:"scale"`
	S     float64 `json:"s"`
}

type FunctionProfile struct {
	Name            string                 `json:"name"`
	Response        []ResponseDistribution `json:"responseDistributions"`
	Replicas        int                    `json:"replicas"`
	ResponsePayload map[string]interface{} `json:"responsePayload"`
	MaxRps          float64                `json:"maxRps"`
	MaxHostRps      float64                `json:"maxHostRps"`
}

type SimulatorDefinition struct {
	Name             string            `json:"name"`
	FunctionProfiles []FunctionProfile `json:"functionProfiles"`
	NetworkLatency   float64           `json:"networkLatency"`
}

func ReadSimulatorDefinition(path string) (SimulatorDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return SimulatorDefinition{}, err
	}
	defer file.Close()
	var simulatorDefinition SimulatorDefinition
	if err := json.NewDecoder(file).Decode(&simulatorDefinition); err != nil {
		return SimulatorDefinition{}, err
	}

	return simulatorDefinition, nil
}
