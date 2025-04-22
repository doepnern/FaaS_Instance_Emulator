package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type ResponseDistribution struct {
	Loc   float64 `json:"loc"`
	Scale float64 `json:"scale"`
	S     float64 `json:"s"`
}

type Constraint struct {
	Resource              string   `json:"resource"`
	Amount                float64  `json:"amount"`
	ResourcePerRequest    float64  `json:"resourcePerRequest"`
	DurationMultiplicator *float64 `json:"durationMultiplicator"`
	AffectsUtilization    bool     `json:"affectsUtilization"`
}

type FunctionProfile struct {
	Name string `json:"name"`
	// response is an arbitrary object
	Response            ResponseDistribution   `json:"responseDistribution"`
	InstanceConstraints []Constraint           `json:"instanceConstraints"`
	Replicas            int                    `json:"replicas"`
	ResponsePayload     map[string]interface{} `json:"responsePayload"`
}

type SimulatorDefinition struct {
	Name              string            `json:"name"`
	FunctionProfiles  []FunctionProfile `json:"functionProfiles"`
	GlobalConstraints []Constraint      `json:"globalConstraints"`
	NetworkLatency    float64           `json:"networkLatency"`
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
	fmt.Println(simulatorDefinition)
	fmt.Println(simulatorDefinition.FunctionProfiles[0].InstanceConstraints[0].Resource)

	return simulatorDefinition, nil
}
