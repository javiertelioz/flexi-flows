package test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/javiertelioz/go-flows/pkg/workflow/config"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

type ConfigTestSuite struct {
	suite.Suite
	jsonFilePath string
	yamlFilePath string
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (suite *ConfigTestSuite) SetupTest() {
	// Create temporary JSON config file
	jsonData := config.WorkflowConfig{
		Nodes: []config.NodeConfig{
			{
				ID:       "node1",
				Type:     "Task",
				TaskFunc: "taskFunc1",
			},
		},
		Edges: []config.EdgeConfig{
			{
				From: "node1",
				To:   "node2",
			},
		},
	}
	jsonBytes, _ := json.Marshal(jsonData)
	suite.jsonFilePath = "test_config.json"

	err := os.WriteFile(suite.jsonFilePath, jsonBytes, 0644)
	if err != nil {
		return
	}

	// Create temporary YAML config file
	yamlData := config.WorkflowConfig{
		Nodes: []config.NodeConfig{
			{
				ID:       "node2",
				Type:     "Task",
				TaskFunc: "taskFunc2",
			},
		},
		Edges: []config.EdgeConfig{
			{
				From: "node2",
				To:   "node3",
			},
		},
	}
	yamlBytes, _ := yaml.Marshal(yamlData)
	suite.yamlFilePath = "test_config.yaml"

	err = os.WriteFile(suite.yamlFilePath, yamlBytes, 0644)
	if err != nil {
		return
	}
}

func (suite *ConfigTestSuite) TearDownTest() {
	os.Remove(suite.jsonFilePath)
	os.Remove(suite.yamlFilePath)
}

func (suite *ConfigTestSuite) TestLoadJSONConfigSuccess() {
	config, err := config.LoadConfig(suite.jsonFilePath)
	suite.NoError(err)
	suite.NotNil(config)
	suite.Equal("node1", config.Nodes[0].ID)
	suite.Equal("Task", config.Nodes[0].Type)
	suite.Equal("taskFunc1", config.Nodes[0].TaskFunc)
}

func (suite *ConfigTestSuite) TestLoadJSONConfigFileNotFound() {
	_, err := config.LoadConfig("non_existent.json")
	suite.Error(err)
}

func (suite *ConfigTestSuite) TestLoadYAMLConfigSuccess() {
	config, err := config.LoadYAMLConfig(suite.yamlFilePath)
	suite.NoError(err)
	suite.NotNil(config)
	suite.Equal("node2", config.Nodes[0].ID)
	suite.Equal("Task", config.Nodes[0].Type)
	suite.Equal("taskFunc2", config.Nodes[0].TaskFunc)
}

func (suite *ConfigTestSuite) TestLoadYAMLConfigFileNotFound() {
	_, err := config.LoadYAMLConfig("non_existent.yaml")
	suite.Error(err)
}

func (suite *ConfigTestSuite) TestLoadJSONConfigInvalidFormat() {
	// Create a file with invalid JSON content
	invalidFilePath := "invalid_config.json"
	os.WriteFile(invalidFilePath, []byte("invalid json"), 0644)
	defer os.Remove(invalidFilePath)

	_, err := config.LoadConfig(invalidFilePath)
	suite.Error(err)
}

func (suite *ConfigTestSuite) TestLoadYAMLConfigInvalidFormat() {
	// Create a file with invalid YAML content
	invalidFilePath := "invalid_config.yaml"
	os.WriteFile(invalidFilePath, []byte("invalid yaml"), 0644)
	defer os.Remove(invalidFilePath)

	_, err := config.LoadYAMLConfig(invalidFilePath)
	suite.Error(err)
}
