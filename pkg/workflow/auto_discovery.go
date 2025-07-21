package workflow

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
)

// FunctionMetadata representa metadata de una función descubierta
type FunctionMetadata struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Timeout     string                 `json:"timeout,omitempty"`
	Retries     int                    `json:"retries,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	ReturnType  string                 `json:"return_type,omitempty"`
	FilePath    string                 `json:"file_path,omitempty"`
}

// AutoDiscoverer maneja el descubrimiento automático de funciones
type AutoDiscoverer struct {
	functions       map[string]interface{}      // funciones descubiertas
	metadata        map[string]FunctionMetadata // metadata de funciones
	fileAnnotations map[string][]string         // anotaciones por archivo
}

// NewAutoDiscoverer crea una nueva instancia del auto-discoverer
func NewAutoDiscoverer() *AutoDiscoverer {
	return &AutoDiscoverer{
		functions:       make(map[string]interface{}),
		metadata:        make(map[string]FunctionMetadata),
		fileAnnotations: make(map[string][]string),
	}
}

// ScanPackage escanea un paquete en busca de funciones válidas
func (ad *AutoDiscoverer) ScanPackage(packagePath string) error {
	// Obtener path absoluto
	absPath, err := filepath.Abs(packagePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Parse del paquete
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, absPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse package: %w", err)
	}

	// Analizar cada archivo en el paquete
	for _, pkg := range packages {
		for filename, file := range pkg.Files {
			err := ad.analyzeFile(fset, filename, file)
			if err != nil {
				return fmt.Errorf("failed to analyze file %s: %w", filename, err)
			}
		}
	}

	return nil
}

// RegisterWithAnnotations registra una función con metadata específicos
func (ad *AutoDiscoverer) RegisterWithAnnotations(name string, fn interface{}, metadata FunctionMetadata) error {
	// Validar la función primero
	if err := ad.ValidateFunction(name, fn); err != nil {
		return err
	}

	// Registrar función y metadata
	ad.functions[name] = fn
	metadata.Name = name
	ad.metadata[name] = metadata

	return nil
}

// ValidateFunction valida que una función tenga la firma correcta para workflows
func (ad *AutoDiscoverer) ValidateFunction(name string, fn interface{}) error {
	fnType := reflect.TypeOf(fn)

	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("invalid function signature for %s: not a function", name)
	}

	numIn := fnType.NumIn()
	numOut := fnType.NumOut()

	// Validar número de parámetros de salida (siempre debe ser 2: result, error)
	if numOut != 2 {
		return fmt.Errorf("invalid function signature for %s: must return 2 values (result, error)", name)
	}

	// Validar que el segundo valor de retorno sea error
	if !fnType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return fmt.Errorf("invalid function signature for %s: second return value must be error", name)
	}

	// Validar parámetros de entrada
	switch numIn {
	case 1:
		// Firma: func(data map[string]interface{}) (map[string]interface{}, error)
		if !ad.isMapStringInterface(fnType.In(0)) {
			return fmt.Errorf("invalid function signature for %s: single parameter must be map[string]interface{}", name)
		}
	case 2:
		// Firma: func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error)
		if !ad.isContext(fnType.In(0)) {
			return fmt.Errorf("invalid function signature for %s: first parameter must be context.Context", name)
		}
		if !ad.isMapStringInterface(fnType.In(1)) {
			return fmt.Errorf("invalid function signature for %s: second parameter must be map[string]interface{}", name)
		}
	default:
		return fmt.Errorf("invalid function signature for %s: must have 1 or 2 parameters", name)
	}

	// Validar primer valor de retorno
	if !ad.isMapStringInterface(fnType.Out(0)) {
		return fmt.Errorf("invalid function signature for %s: first return value must be map[string]interface{}", name)
	}

	return nil
}

// GetDiscoveredFunctions retorna todas las funciones descubiertas
func (ad *AutoDiscoverer) GetDiscoveredFunctions() map[string]interface{} {
	return ad.functions
}

// GetFunctionMetadata retorna los metadata de una función específica
func (ad *AutoDiscoverer) GetFunctionMetadata(name string) *FunctionMetadata {
	if metadata, exists := ad.metadata[name]; exists {
		return &metadata
	}
	return nil
}

// RegisterInManager registra todas las funciones descubiertas en un WorkflowManager
func (ad *AutoDiscoverer) RegisterInManager(manager *WorkflowManager) error {
	for name, fn := range ad.functions {
		// Convertir función a TaskFunc compatible
		taskFunc, err := ad.convertToTaskFunc(fn)
		if err != nil {
			return fmt.Errorf("failed to convert function %s: %w", name, err)
		}

		manager.RegisterTask(name, taskFunc)
	}
	return nil
}

// GenerateWorkflowConfig genera automáticamente una configuración de workflow
func (ad *AutoDiscoverer) GenerateWorkflowConfig(workflowName string) (*config.WorkflowConfig, error) {
	if len(ad.functions) == 0 {
		return nil, fmt.Errorf("no functions discovered to generate workflow")
	}

	workflowConfig := &config.WorkflowConfig{
		Name:      workflowName,
		StartNode: ad.determineStartNode(),
		Nodes:     []config.NodeConfig{},
	}

	// Generar nodos basados en funciones y sus dependencias
	for name, metadata := range ad.metadata {
		node := config.NodeConfig{
			ID:          name,
			Type:        "task",
			Function:    name,
			Description: metadata.Description,
		}

		// Configurar timeout si está especificado
		if metadata.Timeout != "" {
			node.Timeout = metadata.Timeout
		}

		workflowConfig.Nodes = append(workflowConfig.Nodes, node)
	}

	// Configurar navegación basada en DependsOn
	ad.configureNavigation(workflowConfig)

	return workflowConfig, nil
}

// ScanFileForAnnotations escanea un archivo en busca de anotaciones en comentarios
func (ad *AutoDiscoverer) ScanFileForAnnotations(filePath string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	var annotations []string

	// Extraer anotaciones de comentarios
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if strings.HasPrefix(text, "@") {
				annotations = append(annotations, text)
			}
		}
	}

	ad.fileAnnotations[filePath] = annotations
	return nil
}

// GetFileAnnotations retorna las anotaciones encontradas en un archivo
func (ad *AutoDiscoverer) GetFileAnnotations(filePath string) []string {
	return ad.fileAnnotations[filePath]
}

// Métodos auxiliares privados

func (ad *AutoDiscoverer) isMapStringInterface(t reflect.Type) bool {
	return t.Kind() == reflect.Map &&
		t.Key().Kind() == reflect.String &&
		t.Elem().Kind() == reflect.Interface
}

func (ad *AutoDiscoverer) isContext(t reflect.Type) bool {
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	return t.Implements(contextType)
}

func (ad *AutoDiscoverer) analyzeFile(fset *token.FileSet, filename string, file *ast.File) error {
	// Buscar funciones exportadas con firmas válidas
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name.IsExported() {
				// Analizar la función y extraer información
				ad.analyzeFuncDecl(filename, node)
			}
		}
		return true
	})
	return nil
}

func (ad *AutoDiscoverer) analyzeFuncDecl(filename string, funcDecl *ast.FuncDecl) {
	name := funcDecl.Name.Name

	// Crear metadata básico basado en comentarios
	metadata := FunctionMetadata{
		Name:     name,
		FilePath: filename,
		Tags:     []string{"auto-discovered"},
	}

	// Extraer información de comentarios si existen
	if funcDecl.Doc != nil {
		for _, comment := range funcDecl.Doc.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

			// Buscar anotaciones específicas
			if strings.HasPrefix(text, "@description") {
				metadata.Description = strings.TrimSpace(strings.TrimPrefix(text, "@description"))
			} else if strings.HasPrefix(text, "@timeout") {
				metadata.Timeout = strings.TrimSpace(strings.TrimPrefix(text, "@timeout"))
			}
		}
	}

	// Guardar metadata (nota: aquí no tenemos la función real, solo la declaración)
	ad.metadata[name] = metadata
}

func (ad *AutoDiscoverer) convertToTaskFunc(fn interface{}) (interface{}, error) {
	fnType := reflect.TypeOf(fn)

	// Si la función ya tiene la firma correcta, devolverla tal como está
	if ad.isValidTaskFunc(fnType) {
		return fn, nil
	}

	return nil, fmt.Errorf("cannot convert function to TaskFunc")
}

func (ad *AutoDiscoverer) isValidTaskFunc(fnType reflect.Type) bool {
	// Verificar si es una función TaskFunc válida
	numIn := fnType.NumIn()
	numOut := fnType.NumOut()

	return numOut == 2 &&
		fnType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) &&
		((numIn == 1 && ad.isMapStringInterface(fnType.In(0))) ||
			(numIn == 2 && ad.isContext(fnType.In(0)) && ad.isMapStringInterface(fnType.In(1))))
}

func (ad *AutoDiscoverer) determineStartNode() string {
	// Buscar un nodo sin dependencias para ser el nodo inicial
	for name, metadata := range ad.metadata {
		if len(metadata.DependsOn) == 0 {
			return name
		}
	}

	// Si todos tienen dependencias, tomar el primero
	for name := range ad.metadata {
		return name
	}

	return ""
}

func (ad *AutoDiscoverer) configureNavigation(config *config.WorkflowConfig) {
	// Configurar navegación basada en dependencias
	for i, node := range config.Nodes {
		if metadata, exists := ad.metadata[node.ID]; exists {
			// Si no tiene dependencias, podría ser nodo inicial
			if len(metadata.DependsOn) == 0 {
				if config.StartNode == "" || config.StartNode == node.ID {
					config.StartNode = node.ID
				}
			}

			// Configurar next basado en qué nodos dependen de este
			nextNodes := ad.findNextNodes(node.ID)
			if len(nextNodes) > 0 {
				config.Nodes[i].Next = nextNodes
			}
		}
	}
}

func (ad *AutoDiscoverer) findNextNodes(currentNode string) []string {
	var nextNodes []string

	// Buscar nodos que dependan del nodo actual
	for name, metadata := range ad.metadata {
		for _, dep := range metadata.DependsOn {
			if dep == currentNode {
				nextNodes = append(nextNodes, name)
			}
		}
	}

	return nextNodes
}
