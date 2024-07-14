package config

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
)

type NodeConfig struct {
	ID            string   `json:"id" yaml:"id"`
	Type          string   `json:"type" yaml:"type"`
	TaskFunc      string   `json:"taskFunc,omitempty" yaml:"taskFunc,omitempty"`
	ParallelTasks []string `json:"parallelTasks,omitempty" yaml:"parallelTasks,omitempty"`
	BeforeExecute string   `json:"beforeExecute,omitempty" yaml:"beforeExecute,omitempty"`
	AfterExecute  string   `json:"afterExecute,omitempty" yaml:"afterExecute,omitempty"`
}

type EdgeConfig struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type WorkflowConfig struct {
	Nodes []NodeConfig `json:"nodes"`
	Edges []EdgeConfig `json:"edges"`
}

func LoadConfig(filePath string) (*WorkflowConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config WorkflowConfig
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadYAMLConfig(filePath string) (*WorkflowConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config WorkflowConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
