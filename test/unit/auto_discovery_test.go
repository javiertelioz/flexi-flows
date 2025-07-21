package unit

import (
	"context"
	"testing"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
	"github.com/stretchr/testify/suite"
)

// AutoDiscoveryTestSuite agrupa todas las pruebas del sistema de auto-discovery
type AutoDiscoveryTestSuite struct {
	suite.Suite
	manager *workflow.WorkflowManager
}

func TestAutoDiscoveryTestSuite(t *testing.T) {
	suite.Run(t, new(AutoDiscoveryTestSuite))
}

func (suite *AutoDiscoveryTestSuite) SetupTest() {
	suite.manager = workflow.NewWorkflowManager()
}


// TestAutoDiscoveryWithAnnotations prueba el descubrimiento usando anotaciones
func (suite *AutoDiscoveryTestSuite) TestAutoDiscoveryWithAnnotations() {
	// Test: Debe poder leer anotaciones de funciones para auto-configuración
	discoverer := workflow.NewAutoDiscoverer()

	// Ejemplo de función con anotaciones
	testFunc := func(data map[string]interface{}) (map[string]interface{}, error) {
		data["processed"] = true
		return data, nil
	}

	// Registrar función con metadata/anotaciones
	err := discoverer.RegisterWithAnnotations("TestTask", testFunc, workflow.FunctionMetadata{
		Description: "Test task for auto-discovery",
		Timeout:     "30s",
		Retries:     3,
		Tags:        []string{"test", "auto-discovery"},
	})
	suite.NoError(err)

	// Verificar que la función fue registrada con metadata
	metadata := discoverer.GetFunctionMetadata("TestTask")
	suite.NotNil(metadata)
	suite.Equal("Test task for auto-discovery", metadata.Description)
	suite.Equal("30s", metadata.Timeout)
	suite.Equal(3, metadata.Retries)
	suite.Contains(metadata.Tags, "test")
}

// TestAutoDiscoveryFunctionSignatureValidation prueba la validación de firmas de funciones
func (suite *AutoDiscoveryTestSuite) TestAutoDiscoveryFunctionSignatureValidation() {
	discoverer := workflow.NewAutoDiscoverer()

	// Función válida: (map[string]interface{}) (map[string]interface{}, error)
	validFunc := func(data map[string]interface{}) (map[string]interface{}, error) {
		return data, nil
	}

	// Función válida: (context.Context, map[string]interface{}) (map[string]interface{}, error)
	validFuncWithCtx := func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
		return data, nil
	}

	// Función inválida: firma incorrecta
	invalidFunc := func(data string) string {
		return data
	}

	// Test funciones válidas
	err := discoverer.ValidateFunction("ValidFunc", validFunc)
	suite.NoError(err)

	err = discoverer.ValidateFunction("ValidFuncWithCtx", validFuncWithCtx)
	suite.NoError(err)

	// Test función inválida
	err = discoverer.ValidateFunction("InvalidFunc", invalidFunc)
	suite.Error(err)
	suite.Contains(err.Error(), "invalid function signature")
}

// TestAutoConfigGeneration prueba la generación automática de configuración
func (suite *AutoDiscoveryTestSuite) TestAutoConfigGeneration() {
	discoverer := workflow.NewAutoDiscoverer()

	// Registrar algunas funciones con metadata
	task1 := func(data map[string]interface{}) (map[string]interface{}, error) {
		return data, nil
	}

	task2 := func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
		return data, nil
	}

	err := discoverer.RegisterWithAnnotations("Task1", task1, workflow.FunctionMetadata{
		Description: "First task",
		Tags:        []string{"main", "process"},
	})
	suite.NoError(err)

	err = discoverer.RegisterWithAnnotations("Task2", task2, workflow.FunctionMetadata{
		Description: "Second task",
		Tags:        []string{"secondary", "process"},
		DependsOn:   []string{"Task1"},
	})
	suite.NoError(err)

	// Generar configuración automática
	config, err := discoverer.GenerateWorkflowConfig("auto_generated_workflow")
	suite.NoError(err)
	suite.NotNil(config)

	// Verificar configuración generada
	suite.Equal("auto_generated_workflow", config.Name)
	suite.NotEmpty(config.Nodes)
	suite.Len(config.Nodes, 2)

	// Verificar que se respetan las dependencias
	task1Node := findNodeByID(config.Nodes, "Task1")
	task2Node := findNodeByID(config.Nodes, "Task2")

	suite.NotNil(task1Node)
	suite.NotNil(task2Node)

	// Task1 debería apuntar a Task2 basado en DependsOn
	suite.Contains(task1Node.Next, "Task2")
}

// TestAutoDiscoveryWithFileAnnotations prueba el descubrimiento desde anotaciones en archivos
func (suite *AutoDiscoveryTestSuite) TestAutoDiscoveryWithFileAnnotations() {
	// Test: Debe poder leer anotaciones desde comentarios en archivos Go
	discoverer := workflow.NewAutoDiscoverer()

	// Escanear archivos con anotaciones en comentarios
	err := discoverer.ScanFileForAnnotations("../../examples/task_flow/uses_cases/task1.go")
	suite.NoError(err)

	// Verificar que encontró anotaciones
	annotations := discoverer.GetFileAnnotations("../../examples/task_flow/uses_cases/task1.go")
	suite.NotNil(annotations)
}

// Helper function para encontrar nodo por ID
func findNodeByID(nodes []config.NodeConfig, id string) *config.NodeConfig {
	for _, node := range nodes {
		if node.ID == id {
			return &node
		}
	}
	return nil
}
