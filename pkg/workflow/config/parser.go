package config

import (
	"os"

	"encoding/json"

	"gopkg.in/yaml.v3"
)

func LoadConfigFromJSON(filename string) (*WorkflowConfig, error) {
	data, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	var config WorkflowConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadConfigFromYAML(filename string) (*WorkflowConfig, error) {
	data, err := os.ReadFile(filename)
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
