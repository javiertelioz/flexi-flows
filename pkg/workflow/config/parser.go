package config

import (
	"os"

	"encoding/json"

	"gopkg.in/yaml.v3"
)

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
