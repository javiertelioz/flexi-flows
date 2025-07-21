package unit

import (
	"testing"
	"time"

	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
	"github.com/stretchr/testify/suite"
)

// ConfigParserTestSuite agrupa todas las pruebas del sistema de configuración mejorado
type ConfigParserTestSuite struct {
	suite.Suite
	parser *config.ConfigParser
}

func TestConfigParserTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigParserTestSuite))
}

func (suite *ConfigParserTestSuite) SetupTest() {
	suite.parser = config.NewConfigParser()
}

// TestParseSimpleJSONConfig prueba el parsing básico de JSON
func (suite *ConfigParserTestSuite) TestParseSimpleJSONConfig() {
	jsonConfig := `{
		"name": "test_workflow",
		"description": "A test workflow",
		"start_node": "start",
		"nodes": [
			{
				"id": "start",
				"type": "task",
				"function": "StartTask",
				"next": ["end"]
			},
			{
				"id": "end",
				"type": "task",
				"function": "EndTask"
			}
		]
	}`

	config, err := suite.parser.ParseFromJSON([]byte(jsonConfig))
	suite.NoError(err)
	suite.NotNil(config)

	suite.Equal("test_workflow", config.Name)
	suite.Equal("A test workflow", config.Description)
	suite.Equal("start", config.StartNode)
	suite.Len(config.Nodes, 2)

	// Verificar que se expandieron los edges
	suite.Len(config.Edges, 1)
	suite.Equal("start", config.Edges[0].From)
	suite.Equal("end", config.Edges[0].To)
}

// TestParseSimpleYAMLConfig prueba el parsing básico de YAML
func (suite *ConfigParserTestSuite) TestParseSimpleYAMLConfig() {
	yamlConfig := `
name: test_workflow
description: A test workflow  
start_node: start
nodes:
  - id: start
    type: task
    function: StartTask
    next: [end]
  - id: end
    type: task
    function: EndTask
`

	config, err := suite.parser.ParseFromYAML([]byte(yamlConfig))
	suite.NoError(err)
	suite.NotNil(config)

	suite.Equal("test_workflow", config.Name)
	suite.Equal("A test workflow", config.Description)
	suite.Equal("start", config.StartNode)
	suite.Len(config.Nodes, 2)
}

// TestParseConfigWithVariables prueba el sistema de templating
func (suite *ConfigParserTestSuite) TestParseConfigWithVariables() {
	jsonConfig := `{
		"name": "templated_workflow",
		"start_node": "start",
		"variables": {
			"base_url": "https://api.example.com",
			"api_timeout": "30s"
		},
		"nodes": [
			{
				"id": "start",
				"type": "http",
				"url": "${base_url}/users",
				"timeout": "${api_timeout}",
				"method": "GET"
			}
		]
	}`

	config, err := suite.parser.ParseFromJSON([]byte(jsonConfig))
	suite.NoError(err)
	suite.NotNil(config)

	// Verificar que las variables se aplicaron correctamente
	suite.Equal("https://api.example.com/users", config.Nodes[0].URL)
	suite.Equal("30s", config.Nodes[0].Timeout)
	suite.Equal("GET", config.Nodes[0].Method)
}

// TestParseConfigWithComplexNavigation prueba navegación compleja
func (suite *ConfigParserTestSuite) TestParseConfigWithComplexNavigation() {
	jsonConfig := `{
		"name": "complex_workflow",
		"start_node": "start",
		"nodes": [
			{
				"id": "start",
				"type": "task",
				"function": "StartTask",
				"on_success": ["validate"],
				"on_error": ["error_handler"]
			},
			{
				"id": "validate",
				"type": "validation",
				"rules": [
					{
						"field": "email",
						"type": "email",
						"required": true
					}
				],
				"on_success": ["process"],
				"on_error": ["error_handler"]
			},
			{
				"id": "process",
				"type": "task",
				"function": "ProcessData"
			},
			{
				"id": "error_handler",
				"type": "task",
				"function": "HandleError"
			}
		]
	}`

	config, err := suite.parser.ParseFromJSON([]byte(jsonConfig))
	suite.NoError(err)
	suite.NotNil(config)

	suite.Len(config.Nodes, 4)

	// Verificar que se expandieron todos los edges
	suite.Len(config.Edges, 4)

	// Verificar configuración de validación
	validateNode := config.Nodes[1]
	suite.Equal("validate", validateNode.ID)
	suite.Equal("validation", validateNode.Type)
	suite.Len(validateNode.Rules, 1)
	suite.Equal("email", validateNode.Rules[0].Field)
	suite.True(validateNode.Rules[0].Required)
}

// TestParseConfigWithHooks prueba la configuración de hooks
func (suite *ConfigParserTestSuite) TestParseConfigWithHooks() {
	jsonConfig := `{
		"name": "hooks_workflow",
		"start_node": "start",
		"hooks": {
			"audit_log": {
				"type": "after",
				"function": "AuditLog",
				"enabled": true
			}
		},
		"nodes": [
			{
				"id": "start",
				"type": "task",
				"function": "StartTask",
				"hooks": {
					"validation": [{
							"type": "before",
							"function": "ValidateInput"
					}]
				}
			}
		]
	}`

	config, err := suite.parser.ParseFromJSON([]byte(jsonConfig))
	suite.NoError(err)
	suite.NotNil(config)

	// Verificar hooks globales
	suite.Len(config.Hooks, 1)
	auditHook := config.Hooks["audit_log"]
	suite.Equal("after", auditHook.Type)
	suite.Equal("AuditLog", auditHook.Function)
	suite.NotNil(auditHook.Enabled)
	suite.True(*auditHook.Enabled)

	// Verificar hooks del nodo
	nodeHooks := config.Nodes[0].Hooks
	suite.Len(nodeHooks, 1)
	validationHook := nodeHooks["validation"]
	suite.Equal("before", validationHook[0].Type)
	suite.Equal("ValidateInput", validationHook[0].Function)
}

// TestParseConfigWithSettings prueba configuraciones globales
func (suite *ConfigParserTestSuite) TestParseConfigWithSettings() {
	jsonConfig := `{
		"name": "settings_workflow",
		"start_node": "start",
		"settings": {
			"timeout": "5m",
			"max_retries": 5,
			"parallel_limit": 10,
			"enable_debug": true,
			"enable_metrics": true,
			"enable_audit_log": false
		},
		"nodes": [
			{
				"id": "start",
				"type": "task",
				"function": "StartTask"
			}
		]
	}`

	config, err := suite.parser.ParseFromJSON([]byte(jsonConfig))
	suite.NoError(err)
	suite.NotNil(config)

	settings := config.Settings
	suite.Equal("5m", settings.Timeout)
	suite.Equal(5, settings.MaxRetries)
	suite.Equal(10, settings.ParallelLimit)
	suite.True(settings.EnableDebug)
	suite.True(settings.EnableMetrics)
	suite.False(settings.EnableAuditLog)
}

// TestValidationErrors prueba los errores de validación
func (suite *ConfigParserTestSuite) TestValidationErrors() {
	testCases := []struct {
		name     string
		config   string
		errorMsg string
	}{
		{
			name: "missing name",
			config: `{
				"start_node": "start",
				"nodes": [{"id": "start", "type": "task", "function": "Test"}]
			}`,
			errorMsg: "workflow name is required",
		},
		{
			name: "missing start_node",
			config: `{
				"name": "test",
				"nodes": [{"id": "start", "type": "task", "function": "Test"}]
			}`,
			errorMsg: "start_node is required",
		},
		{
			name: "empty nodes",
			config: `{
				"name": "test",
				"start_node": "start",
				"nodes": []
			}`,
			errorMsg: "at least one node is required",
		},
		{
			name: "duplicate node IDs",
			config: `{
				"name": "test",
				"start_node": "start",
				"nodes": [
					{"id": "start", "type": "task", "function": "Test1"},
					{"id": "start", "type": "task", "function": "Test2"}
				]
			}`,
			errorMsg: "duplicate node ID: start",
		},
		{
			name: "start node not found",
			config: `{
				"name": "test",
				"start_node": "missing",
				"nodes": [{"id": "start", "type": "task", "function": "Test"}]
			}`,
			errorMsg: "start node missing not found in nodes",
		},
		{
			name: "task without function",
			config: `{
				"name": "test",
				"start_node": "start",
				"nodes": [{"id": "start", "type": "task"}]
			}`,
			errorMsg: "task nodes require a function",
		},
		{
			name: "http without url",
			config: `{
				"name": "test",
				"start_node": "start",
				"nodes": [{"id": "start", "type": "http"}]
			}`,
			errorMsg: "HTTP nodes require a URL",
		},
		{
			name: "delay without duration",
			config: `{
				"name": "test",
				"start_node": "start",
				"nodes": [{"id": "start", "type": "delay"}]
			}`,
			errorMsg: "delay nodes require a duration",
		},
		{
			name: "delay with invalid duration",
			config: `{
				"name": "test",
				"start_node": "start",
				"nodes": [{"id": "start", "type": "delay", "duration": "invalid"}]
			}`,
			errorMsg: "invalid duration format: invalid",
		},
		{
			name: "conditional without condition",
			config: `{
				"name": "test",
				"start_node": "start",
				"nodes": [{"id": "start", "type": "conditional"}]
			}`,
			errorMsg: "conditional nodes require a condition",
		},
		{
			name: "loop without collection",
			config: `{
				"name": "test",
				"start_node": "start",
				"nodes": [{"id": "start", "type": "loop"}]
			}`,
			errorMsg: "loop nodes require a collection",
		},
		{
			name: "validation without rules",
			config: `{
				"name": "test",
				"start_node": "start",
				"nodes": [{"id": "start", "type": "validation"}]
			}`,
			errorMsg: "validation nodes require at least one rule",
		},
		{
			name: "parallel without tasks",
			config: `{
				"name": "test",
				"start_node": "start",
				"nodes": [{"id": "start", "type": "parallel"}]
			}`,
			errorMsg: "parallel nodes require parallel tasks",
		},
		{
			name: "invalid navigation reference",
			config: `{
				"name": "test",
				"start_node": "start",
				"nodes": [
					{
						"id": "start",
						"type": "task",
						"function": "Test",
						"next": ["missing_node"]
					}
				]
			}`,
			errorMsg: "node start references non-existent node missing_node in next",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.parser.ParseFromJSON([]byte(tc.config))
			suite.Error(err)
			suite.Contains(err.Error(), tc.errorMsg)
		})
	}
}

// TestTemplateEngine prueba el motor de templates por separado
func (suite *ConfigParserTestSuite) TestTemplateEngine() {
	engine := config.NewTemplateEngine()

	variables := map[string]interface{}{
		"api_url": "https://api.example.com",
		"version": "v1",
		"port":    8080,
	}

	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "URL: ${api_url}/${version}",
			expected: "URL: https://api.example.com/v1",
		},
		{
			input:    "Port: ${port}",
			expected: "Port: 8080",
		},
		{
			input:    "No variables here",
			expected: "No variables here",
		},
		{
			input:    "${api_url} and ${missing_var}",
			expected: "https://api.example.com and ${missing_var}",
		},
	}

	for _, tc := range testCases {
		result, err := engine.Process(tc.input, variables)
		suite.NoError(err)
		suite.Equal(tc.expected, result)
	}
}

// TestNodeConfigTypes prueba configuraciones específicas de tipos de nodos
func (suite *ConfigParserTestSuite) TestNodeConfigTypes() {
	jsonConfig := `{
		"name": "node_types_workflow",
		"start_node": "http_node",
		"nodes": [
			{
				"id": "http_node",
				"type": "http",
				"url": "https://api.example.com/data",
				"method": "POST",
				"headers": {
					"Content-Type": "application/json",
					"Authorization": "Bearer token123"
				},
				"body": {
					"key": "value"
				},
				"next": ["delay_node"]
			},
			{
				"id": "delay_node",
				"type": "delay",
				"duration": "5s",
				"next": ["validation_node"]
			},
			{
				"id": "validation_node",
				"type": "validation",
				"rules": [
					{
						"field": "email",
						"type": "email",
						"required": true,
						"message": "Invalid email format"
					},
					{
						"field": "age",
						"type": "number",
						"min": 0,
						"max": 120
					}
				],
				"next": ["transform_node"]
			},
			{
				"id": "transform_node",
				"type": "transform",
				"mapping": {
					"user_email": "data.email",
					"user_age": "data.age"
				}
			}
		]
	}`

	config, err := suite.parser.ParseFromJSON([]byte(jsonConfig))
	suite.NoError(err)
	suite.NotNil(config)

	// Verificar nodo HTTP
	httpNode := config.Nodes[0]
	suite.Equal("http_node", httpNode.ID)
	suite.Equal("https://api.example.com/data", httpNode.URL)
	suite.Equal("POST", httpNode.Method)
	suite.Len(httpNode.Headers, 2)
	suite.Equal("application/json", httpNode.Headers["Content-Type"])
	suite.NotNil(httpNode.Body)

	// Verificar nodo delay
	delayNode := config.Nodes[1]
	suite.Equal("delay_node", delayNode.ID)
	suite.Equal("5s", delayNode.Duration)

	// Verificar parsing de duración
	duration, err := time.ParseDuration(delayNode.Duration)
	suite.NoError(err)
	suite.Equal(5*time.Second, duration)

	// Verificar nodo de validación
	validationNode := config.Nodes[2]
	suite.Equal("validation_node", validationNode.ID)
	suite.Len(validationNode.Rules, 2)

	emailRule := validationNode.Rules[0]
	suite.Equal("email", emailRule.Field)
	suite.Equal("email", emailRule.Type)
	suite.True(emailRule.Required)
	suite.Equal("Invalid email format", emailRule.Message)

	ageRule := validationNode.Rules[1]
	suite.Equal("age", ageRule.Field)
	suite.Equal("number", ageRule.Type)
	suite.Equal(float64(0), ageRule.MinValue)
	suite.Equal(float64(120), ageRule.MaxValue)

	// Verificar nodo de transformación
	transformNode := config.Nodes[3]
	suite.Equal("transform_node", transformNode.ID)
	suite.NotNil(transformNode.Mapping)
	suite.Equal("data.email", transformNode.Mapping["user_email"])
	suite.Equal("data.age", transformNode.Mapping["user_age"])
}
