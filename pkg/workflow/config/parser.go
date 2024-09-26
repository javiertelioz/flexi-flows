package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	JSON = "json"
	YAML = "yaml"
	YML  = "yml"
)

func LoadConfig(filePath string) (*WorkflowConfig, error) {
	fileType := strings.ToLower(
		strings.TrimPrefix(filepath.Ext(filePath), "."),
	)

	switch fileType {
	case JSON:
		return LoadJSONConfig(filePath)
	case YAML, YML:
		return LoadYAMLConfig(filePath)
	default:
		return nil, errors.New("unsupported file type")
	}
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

func LoadJSONConfig(filePath string) (*WorkflowConfig, error) {
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
