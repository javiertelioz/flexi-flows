package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidateTypes valida los tipos de datos según restricciones definidas
func (av *AdvancedValidator) ValidateTypes(ctx context.Context, data interface{}, schema map[string]TypeConstraint) (*AdvancedValidationResult, error) {
	result := &AdvancedValidationResult{
		Valid:   true,
		Errors:  []ValidationError{},
		Data:    data,
		Summary: make(map[string]interface{}),
	}

	// Convertir datos a map para facilitar acceso
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		// Intentar convertir usando JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "root",
				Rule:    "type_conversion",
				Message: "failed to convert data to map for validation",
				Value:   data,
			})
			return result, nil
		}

		if err := json.Unmarshal(jsonData, &dataMap); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "root",
				Rule:    "type_conversion",
				Message: "failed to unmarshal data for validation",
				Value:   data,
			})
			return result, nil
		}
	}

	// Validar cada campo según su esquema
	for fieldName, constraint := range schema {
		// Verificar cancelación de contexto
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		value, exists := dataMap[fieldName]

		// Validar campo requerido
		if constraint.Required && (!exists || value == nil) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Rule:    "required",
				Message: fmt.Sprintf("field '%s' is required", fieldName),
				Value:   nil,
			})
			continue
		}

		if !exists || value == nil {
			continue // Campo opcional no presente
		}

		// Validar tipo
		if err := av.validateFieldType(fieldName, value, constraint, result); err != nil {
			// Error ya agregado a result.Errors
		}
	}

	// Preparar resumen
	result.Summary["total_fields"] = len(schema)
	result.Summary["errors_count"] = len(result.Errors)
	result.Summary["validation_time"] = time.Now()

	return result, nil
}

// ValidateDependencies valida las dependencias entre nodos
func (av *AdvancedValidator) ValidateDependencies(ctx context.Context, nodes []NodeConfig) (*AdvancedValidationResult, error) {
	result := &AdvancedValidationResult{
		Valid:   true,
		Errors:  []ValidationError{},
		Data:    nodes,
		Summary: make(map[string]interface{}),
	}

	// Crear mapa de nodos para búsqueda rápida
	nodeMap := make(map[string]NodeConfig)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}

	// Validar dependencias
	for _, node := range nodes {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		for _, depID := range node.Dependencies {
			// Verificar que la dependencia existe
			if _, exists := nodeMap[depID]; !exists {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   node.ID,
					Rule:    "dependency_not_found",
					Message: fmt.Sprintf("node '%s' depends on non-existent node '%s'", node.ID, depID),
					Value:   depID,
				})
			}
		}
	}

	// Detectar dependencias circulares
	if cycles := av.detectDependencyCycles(nodes); len(cycles) > 0 {
		result.Valid = false
		for _, cycle := range cycles {
			result.Errors = append(result.Errors, ValidationError{
				Field:   strings.Join(cycle, "->"),
				Rule:    "circular_dependency",
				Message: fmt.Sprintf("circular dependency detected: %s", strings.Join(cycle, " -> ")),
				Value:   cycle,
			})
		}
	}

	result.Summary["total_nodes"] = len(nodes)
	result.Summary["errors_count"] = len(result.Errors)

	return result, nil
}

// DetectCycles detecta ciclos en un grafo de nodos
func (av *AdvancedValidator) DetectCycles(ctx context.Context, graph GraphDefinition) (*AdvancedValidationResult, error) {
	result := &AdvancedValidationResult{
		Valid:   true,
		Errors:  []ValidationError{},
		Data:    graph,
		Summary: make(map[string]interface{}),
		Cycles:  [][]string{},
	}

	// Crear mapa de adyacencia usando SOLO las Edges del grafo
	// No mezclamos Dependencies con Edges ya que representan direcciones opuestas
	adjList := make(map[string][]string)
	nodeSet := make(map[string]bool)

	// Registrar todos los nodos
	for _, node := range graph.Nodes {
		nodeSet[node.ID] = true
		if adjList[node.ID] == nil {
			adjList[node.ID] = []string{}
		}
	}

	// Construir lista de adyacencia SOLO desde las Edges
	for _, edge := range graph.Edges {
		if _, exists := adjList[edge.From]; !exists {
			adjList[edge.From] = []string{}
		}
		adjList[edge.From] = append(adjList[edge.From], edge.To)
		// Asegurar que el nodo 'To' también existe en el mapa
		if _, exists := adjList[edge.To]; !exists {
			adjList[edge.To] = []string{}
		}
	}

	// Detectar ciclos usando DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeID := range nodeSet {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		if !visited[nodeID] {
			path := []string{}
			cycles := av.dfsCycleDetection(nodeID, adjList, visited, recStack, path)
			result.Cycles = append(result.Cycles, cycles...)
		}
	}

	if len(result.Cycles) > 0 {
		result.Valid = false
		for _, cycle := range result.Cycles {
			result.Errors = append(result.Errors, ValidationError{
				Field:   strings.Join(cycle, "->"),
				Rule:    "cycle_detected",
				Message: fmt.Sprintf("cycle detected in graph: %s", strings.Join(cycle, " -> ")),
				Value:   cycle,
			})
		}
	}

	result.Summary["total_nodes"] = len(graph.Nodes)
	result.Summary["cycles_found"] = len(result.Cycles)
	result.Summary["errors_count"] = len(result.Errors)

	return result, nil
}

// ValidateFunctionSignature valida la firma de una función
func (av *AdvancedValidator) ValidateFunctionSignature(ctx context.Context, fn interface{}) (*AdvancedValidationResult, error) {
	result := &AdvancedValidationResult{
		Valid:   true,
		Errors:  []ValidationError{},
		Data:    nil, // No podemos serializar funciones
		Summary: make(map[string]interface{}),
	}

	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "function",
			Rule:    "not_a_function",
			Message: "provided value is not a function",
			Value:   fnType.String(),
		})
		return result, nil
	}

	// Verificar número de parámetros (mínimo 2: context.Context, interface{})
	if fnType.NumIn() < 2 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "function",
			Rule:    "invalid_parameters",
			Message: "function must have at least 2 parameters: (context.Context, interface{})",
			Value:   fnType.NumIn(),
		})
	}

	// Verificar que el primer parámetro es context.Context
	if fnType.NumIn() > 0 {
		firstParam := fnType.In(0)
		contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
		if !firstParam.Implements(contextType) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "function",
				Rule:    "invalid_first_parameter",
				Message: "first parameter must be context.Context",
				Value:   firstParam.String(),
			})
		}
	}

	// Verificar número de valores de retorno (debe ser 2: interface{}, error)
	if fnType.NumOut() != 2 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "function",
			Rule:    "invalid_returns",
			Message: "function must return (interface{}, error)",
			Value:   fnType.NumOut(),
		})
	}

	// Verificar que el segundo valor de retorno es error
	if fnType.NumOut() > 1 {
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		secondReturn := fnType.Out(1)
		if !secondReturn.Implements(errorType) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "function",
				Rule:    "invalid_second_return",
				Message: "second return value must be error",
				Value:   secondReturn.String(),
			})
		}
	}

	result.Summary["parameter_count"] = fnType.NumIn()
	result.Summary["return_count"] = fnType.NumOut()
	result.Summary["errors_count"] = len(result.Errors)

	return result, nil
}

// ValidateComplexSchema valida datos contra un esquema complejo
func (av *AdvancedValidator) ValidateComplexSchema(ctx context.Context, data interface{}, schema ComplexSchema) (*AdvancedValidationResult, error) {
	result := &AdvancedValidationResult{
		Valid:   true,
		Errors:  []ValidationError{},
		Data:    data,
		Summary: make(map[string]interface{}),
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "root",
			Rule:    "invalid_type",
			Message: "data must be an object",
			Value:   reflect.TypeOf(data).String(),
		})
		return result, nil
	}

	// Validar cada propiedad del esquema
	for propName, propSchema := range schema.Properties {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		value, exists := dataMap[propName]

		// Verificar campos requeridos
		if propSchema.Required && (!exists || value == nil) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   propName,
				Rule:    "required",
				Message: fmt.Sprintf("property '%s' is required", propName),
				Value:   nil,
			})
			continue
		}

		if !exists || value == nil {
			continue
		}

		// Validar la propiedad recursivamente
		if err := av.validateProperty(propName, value, propSchema, result); err != nil {
			// Error ya agregado a result.Errors por validateProperty
			result.Valid = false
		}
	}

	// Si hay errores, marcar el resultado como inválido
	if len(result.Errors) > 0 {
		result.Valid = false
	}

	result.Summary["total_properties"] = len(schema.Properties)
	result.Summary["errors_count"] = len(result.Errors)

	return result, nil
}

// ValidatePerformanceConstraints valida restricciones de rendimiento
func (av *AdvancedValidator) ValidatePerformanceConstraints(ctx context.Context, data interface{}) (*AdvancedValidationResult, error) {
	result := &AdvancedValidationResult{
		Valid:   true,
		Errors:  []ValidationError{},
		Data:    data,
		Summary: make(map[string]interface{}),
	}

	startTime := time.Now()

	// Convertir datos a map para facilitar acceso
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		// Intentar convertir usando JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "root",
				Rule:    "type_conversion",
				Message: "failed to convert data to map for validation",
				Value:   data,
			})
			return result, nil
		}

		if err := json.Unmarshal(jsonData, &dataMap); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "root",
				Rule:    "type_conversion",
				Message: "failed to unmarshal data for validation",
				Value:   data,
			})
			return result, nil
		}
	}

	// Validar execution_time
	if execTime, exists := dataMap["execution_time"]; exists {
		if execTimeInt, ok := execTime.(int); ok {
			if execTimeInt > 5000 { // Límite de 5 segundos en milisegundos
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   "execution_time",
					Rule:    "max_execution_time_exceeded",
					Message: fmt.Sprintf("execution time (%d ms) exceeds maximum allowed (5000 ms)", execTimeInt),
					Value:   execTimeInt,
				})
			}
		}
	}

	// Validar memory_usage
	if memUsage, exists := dataMap["memory_usage"]; exists {
		if memUsageInt, ok := memUsage.(int); ok {
			if memUsageInt > 1024 { // Límite de 1024 MB
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   "memory_usage",
					Rule:    "max_memory_exceeded",
					Message: fmt.Sprintf("memory usage (%d MB) exceeds maximum allowed (1024 MB)", memUsageInt),
					Value:   memUsageInt,
				})
			}
		}
	}

	// Validar cpu_usage
	if cpuUsage, exists := dataMap["cpu_usage"]; exists {
		if cpuUsageFloat, ok := cpuUsage.(float64); ok {
			if cpuUsageFloat > 100.0 || cpuUsageFloat < 0.0 {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   "cpu_usage",
					Rule:    "invalid_cpu_percentage",
					Message: fmt.Sprintf("CPU usage (%.2f%%) must be between 0 and 100", cpuUsageFloat),
					Value:   cpuUsageFloat,
				})
			}
		}
	}

	// Validar timeout
	if timeout, exists := dataMap["timeout"]; exists {
		if timeoutStr, ok := timeout.(string); ok {
			if _, err := time.ParseDuration(timeoutStr); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   "timeout",
					Rule:    "invalid_duration_format",
					Message: fmt.Sprintf("timeout '%s' is not a valid duration format", timeoutStr),
					Value:   timeoutStr,
				})
			}
		}
	}

	// Verificar tamaño de datos si se especifica límite
	if av.MaxDataSize > 0 {
		jsonData, err := json.Marshal(data)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "data",
				Rule:    "serialization_failed",
				Message: "failed to serialize data for size check",
				Value:   err.Error(),
			})
		} else {
			dataSize := len(jsonData)
			if av.MaxDataSize > 0 && dataSize > av.MaxDataSize {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   "data",
					Rule:    "size_limit_exceeded",
					Message: fmt.Sprintf("data size (%d bytes) exceeds limit (%d bytes)", dataSize, av.MaxDataSize),
					Value:   dataSize,
				})
			}
			result.Summary["data_size_bytes"] = dataSize
		}
	}

	// Verificar tiempo de validación
	validationTime := time.Since(startTime)
	if av.MaxValidationTime > 0 && validationTime > av.MaxValidationTime {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "validation",
			Rule:    "time_limit_exceeded",
			Message: fmt.Sprintf("validation time (%v) exceeds limit (%v)", validationTime, av.MaxValidationTime),
			Value:   validationTime.String(),
		})
	}

	result.Summary["validation_duration"] = validationTime.String()
	result.Summary["errors_count"] = len(result.Errors)

	return result, nil
}

// Métodos helper privados

func (av *AdvancedValidator) validateFieldType(fieldName string, value interface{}, constraint TypeConstraint, result *AdvancedValidationResult) error {
	switch constraint.Type {
	case "string":
		return av.validateStringField(fieldName, value, constraint, result)
	case "int", "integer":
		return av.validateIntField(fieldName, value, constraint, result)
	case "float", "number":
		return av.validateFloatField(fieldName, value, constraint, result)
	case "bool", "boolean":
		return av.validateBoolField(fieldName, value, constraint, result)
	case "array":
		return av.validateArrayField(fieldName, value, constraint, result)
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fieldName,
			Rule:    "unknown_type",
			Message: fmt.Sprintf("unknown type constraint: %s", constraint.Type),
			Value:   constraint.Type,
		})
		return fmt.Errorf("unknown type: %s", constraint.Type)
	}
}

func (av *AdvancedValidator) validateStringField(fieldName string, value interface{}, constraint TypeConstraint, result *AdvancedValidationResult) error {
	strVal, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fieldName,
			Rule:    "type_mismatch",
			Message: fmt.Sprintf("field '%s' must be a string, got %T", fieldName, value),
			Value:   value,
		})
		return fmt.Errorf("type mismatch")
	}

	// Validar longitud mínima
	if constraint.MinLength > 0 && len(strVal) < constraint.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fieldName,
			Rule:    "min_length",
			Message: fmt.Sprintf("field '%s' must be at least %d characters long", fieldName, constraint.MinLength),
			Value:   len(strVal),
		})
	}

	// Validar longitud máxima
	if constraint.MaxLength > 0 && len(strVal) > constraint.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fieldName,
			Rule:    "max_length",
			Message: fmt.Sprintf("field '%s' must be at most %d characters long", fieldName, constraint.MaxLength),
			Value:   len(strVal),
		})
	}

	// Validar patrón
	if constraint.Pattern != "" {
		if matched, err := regexp.MatchString(constraint.Pattern, strVal); err != nil || !matched {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Rule:    "pattern_mismatch",
				Message: fmt.Sprintf("field '%s' does not match required pattern", fieldName),
				Value:   strVal,
			})
		}
	}

	// Validar enum
	if len(constraint.Enum) > 0 {
		found := false
		for _, enumVal := range constraint.Enum {
			if enumVal == strVal {
				found = true
				break
			}
		}
		if !found {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Rule:    "enum_mismatch",
				Message: fmt.Sprintf("field '%s' must be one of: %v", fieldName, constraint.Enum),
				Value:   strVal,
			})
		}
	}

	return nil
}

func (av *AdvancedValidator) validateIntField(fieldName string, value interface{}, constraint TypeConstraint, result *AdvancedValidationResult) error {
	var intVal int64

	switch v := value.(type) {
	case int:
		intVal = int64(v)
	case int64:
		intVal = v
	case float64:
		intVal = int64(v)
	case string:
		var err error
		intVal, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Rule:    "type_conversion",
				Message: fmt.Sprintf("field '%s' cannot be converted to integer", fieldName),
				Value:   value,
			})
			return err
		}
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fieldName,
			Rule:    "type_mismatch",
			Message: fmt.Sprintf("field '%s' must be an integer, got %T", fieldName, value),
			Value:   value,
		})
		return fmt.Errorf("type mismatch")
	}

	// Validar rango mínimo
	if constraint.Min != nil {
		var minVal int64
		switch v := constraint.Min.(type) {
		case int:
			minVal = int64(v)
		case int64:
			minVal = v
		case float64:
			minVal = int64(v)
		default:
			// Si no puede convertir, saltar validación
			goto validateMax
		}

		if intVal < minVal {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Rule:    "min_value",
				Message: fmt.Sprintf("field '%s' must be at least %v", fieldName, constraint.Min),
				Value:   intVal,
			})
		}
	}

validateMax:
	// Validar rango máximo
	if constraint.Max != nil {
		var maxVal int64
		switch v := constraint.Max.(type) {
		case int:
			maxVal = int64(v)
		case int64:
			maxVal = v
		case float64:
			maxVal = int64(v)
		default:
			// Si no puede convertir, saltar validación
			return nil
		}

		if intVal > maxVal {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Rule:    "max_value",
				Message: fmt.Sprintf("field '%s' must be at most %v", fieldName, constraint.Max),
				Value:   intVal,
			})
		}
	}

	return nil
}

func (av *AdvancedValidator) validateFloatField(fieldName string, value interface{}, constraint TypeConstraint, result *AdvancedValidationResult) error {
	var floatVal float64

	switch v := value.(type) {
	case float64:
		floatVal = v
	case int:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	case string:
		var err error
		floatVal, err = strconv.ParseFloat(v, 64)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Rule:    "type_conversion",
				Message: fmt.Sprintf("field '%s' cannot be converted to float", fieldName),
				Value:   value,
			})
			return err
		}
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fieldName,
			Rule:    "type_mismatch",
			Message: fmt.Sprintf("field '%s' must be a number, got %T", fieldName, value),
			Value:   value,
		})
		return fmt.Errorf("type mismatch")
	}

	// Validar rango mínimo
	if constraint.Min != nil {
		if minVal, ok := constraint.Min.(float64); ok && floatVal < minVal {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Rule:    "min_value",
				Message: fmt.Sprintf("field '%s' must be at least %v", fieldName, constraint.Min),
				Value:   floatVal,
			})
		}
	}

	// Validar rango máximo
	if constraint.Max != nil {
		if maxVal, ok := constraint.Max.(float64); ok && floatVal > maxVal {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Rule:    "max_value",
				Message: fmt.Sprintf("field '%s' must be at most %v", fieldName, constraint.Max),
				Value:   floatVal,
			})
		}
	}

	return nil
}

func (av *AdvancedValidator) validateBoolField(fieldName string, value interface{}, constraint TypeConstraint, result *AdvancedValidationResult) error {
	_, ok := value.(bool)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fieldName,
			Rule:    "type_mismatch",
			Message: fmt.Sprintf("field '%s' must be a boolean, got %T", fieldName, value),
			Value:   value,
		})
		return fmt.Errorf("type mismatch")
	}

	return nil
}

func (av *AdvancedValidator) validateArrayField(fieldName string, value interface{}, constraint TypeConstraint, result *AdvancedValidationResult) error {
	arrayVal := reflect.ValueOf(value)
	if arrayVal.Kind() != reflect.Slice && arrayVal.Kind() != reflect.Array {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fieldName,
			Rule:    "type_mismatch",
			Message: fmt.Sprintf("field '%s' must be an array, got %T", fieldName, value),
			Value:   value,
		})
		return fmt.Errorf("type mismatch")
	}

	arrayLen := arrayVal.Len()

	// Validar longitud mínima del array
	if constraint.Min != nil {
		if minLen, ok := constraint.Min.(float64); ok && arrayLen < int(minLen) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Rule:    "min_items",
				Message: fmt.Sprintf("field '%s' must have at least %v items", fieldName, constraint.Min),
				Value:   arrayLen,
			})
		}
	}

	// Validar longitud máxima del array
	if constraint.Max != nil {
		if maxLen, ok := constraint.Max.(float64); ok && arrayLen > int(maxLen) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Rule:    "max_items",
				Message: fmt.Sprintf("field '%s' must have at most %v items", fieldName, constraint.Max),
				Value:   arrayLen,
			})
		}
	}

	return nil
}

func (av *AdvancedValidator) detectDependencyCycles(nodes []NodeConfig) [][]string {
	// Crear mapa de adyacencia para dependencias
	adjList := make(map[string][]string)
	for _, node := range nodes {
		adjList[node.ID] = node.Dependencies
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycles [][]string

	for _, node := range nodes {
		if !visited[node.ID] {
			path := []string{}
			cyclePaths := av.dfsCycleDetection(node.ID, adjList, visited, recStack, path)
			cycles = append(cycles, cyclePaths...)
		}
	}

	return cycles
}

func (av *AdvancedValidator) dfsCycleDetection(nodeID string, adjList map[string][]string, visited, recStack map[string]bool, path []string) [][]string {
	visited[nodeID] = true
	recStack[nodeID] = true
	path = append(path, nodeID)

	var cycles [][]string

	// Explorar vecinos
	for _, neighbor := range adjList[nodeID] {
		if !visited[neighbor] {
			// Recursión DFS
			subCycles := av.dfsCycleDetection(neighbor, adjList, visited, recStack, path)
			cycles = append(cycles, subCycles...)
		} else if recStack[neighbor] {
			// Encontramos un ciclo
			cycleStart := -1
			for i, node := range path {
				if node == neighbor {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append(path[cycleStart:], neighbor)
				cycles = append(cycles, cycle)
			}
		}
	}

	recStack[nodeID] = false
	return cycles
}

func (av *AdvancedValidator) validateProperty(propName string, value interface{}, schema PropertySchema, result *AdvancedValidationResult) error {
	switch schema.Type {
	case "object":
		return av.validateObjectProperty(propName, value, schema, result)
	case "array":
		return av.validateArrayProperty(propName, value, schema, result)
	case "string":
		return av.validateStringProperty(propName, value, schema, result)
	case "int", "integer":
		return av.validateIntProperty(propName, value, schema, result)
	case "number", "float":
		return av.validateNumberProperty(propName, value, schema, result)
	case "bool", "boolean":
		return av.validateBoolProperty(propName, value, schema, result)
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   propName,
			Rule:    "unknown_type",
			Message: fmt.Sprintf("unknown property type: %s", schema.Type),
			Value:   schema.Type,
		})
		return fmt.Errorf("unknown type: %s", schema.Type)
	}
}

func (av *AdvancedValidator) validateObjectProperty(propName string, value interface{}, schema PropertySchema, result *AdvancedValidationResult) error {
	objMap, ok := value.(map[string]interface{})
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   propName,
			Rule:    "type_mismatch",
			Message: fmt.Sprintf("property '%s' must be an object", propName),
			Value:   reflect.TypeOf(value).String(),
		})
		return fmt.Errorf("type mismatch")
	}

	// Validar propiedades anidadas
	for subPropName, subSchema := range schema.Properties {
		subValue, exists := objMap[subPropName]

		if subSchema.Required && (!exists || subValue == nil) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("%s.%s", propName, subPropName),
				Rule:    "required",
				Message: fmt.Sprintf("property '%s.%s' is required", propName, subPropName),
				Value:   nil,
			})
			continue
		}

		if exists && subValue != nil {
			av.validateProperty(fmt.Sprintf("%s.%s", propName, subPropName), subValue, subSchema, result)
		}
	}

	return nil
}

func (av *AdvancedValidator) validateArrayProperty(propName string, value interface{}, schema PropertySchema, result *AdvancedValidationResult) error {
	arrayVal := reflect.ValueOf(value)
	if arrayVal.Kind() != reflect.Slice && arrayVal.Kind() != reflect.Array {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   propName,
			Rule:    "type_mismatch",
			Message: fmt.Sprintf("property '%s' must be an array", propName),
			Value:   reflect.TypeOf(value).String(),
		})
		return fmt.Errorf("type mismatch")
	}

	arrayLen := arrayVal.Len()

	// Validar límites de elementos
	if schema.MinItems > 0 && arrayLen < schema.MinItems {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   propName,
			Rule:    "min_items",
			Message: fmt.Sprintf("property '%s' must have at least %d items", propName, schema.MinItems),
			Value:   arrayLen,
		})
	}

	if schema.MaxItems > 0 && arrayLen > schema.MaxItems {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   propName,
			Rule:    "max_items",
			Message: fmt.Sprintf("property '%s' must have at most %d items", propName, schema.MaxItems),
			Value:   arrayLen,
		})
	}

	// Validar elementos del array si se especifica el esquema de items
	if schema.Items != nil {
		arraySlice, ok := value.([]interface{})
		if !ok {
			// Intentar convertir
			for i := 0; i < arrayLen; i++ {
				elem := arrayVal.Index(i).Interface()
				if err := av.validateProperty(fmt.Sprintf("%s[%d]", propName, i), elem, *schema.Items, result); err != nil {
					// Error ya agregado a result
				}
			}
		} else {
			for i, elem := range arraySlice {
				if err := av.validateProperty(fmt.Sprintf("%s[%d]", propName, i), elem, *schema.Items, result); err != nil {
					// Error ya agregado a result
				}
			}
		}
	}

	return nil
}

func (av *AdvancedValidator) validateStringProperty(propName string, value interface{}, schema PropertySchema, result *AdvancedValidationResult) error {
	strVal, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   propName,
			Rule:    "type_mismatch",
			Message: fmt.Sprintf("property '%s' must be a string", propName),
			Value:   reflect.TypeOf(value).String(),
		})
		return fmt.Errorf("type mismatch")
	}

	strLen := len(strVal)

	// Validar longitud
	if schema.MinLength > 0 && strLen < schema.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   propName,
			Rule:    "min_length",
			Message: fmt.Sprintf("property '%s' must be at least %d characters long", propName, schema.MinLength),
			Value:   strLen,
		})
	}

	if schema.MaxLength > 0 && strLen > schema.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   propName,
			Rule:    "max_length",
			Message: fmt.Sprintf("property '%s' must be at most %d characters long", propName, schema.MaxLength),
			Value:   strLen,
		})
	}

	// Validar patrón
	if schema.Pattern != "" {
		if matched, err := regexp.MatchString(schema.Pattern, strVal); err != nil || !matched {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   propName,
				Rule:    "pattern_mismatch",
				Message: fmt.Sprintf("property '%s' does not match required pattern", propName),
				Value:   strVal,
			})
		}
	}

	// Validar enum
	if len(schema.Enum) > 0 {
		found := false
		for _, enumVal := range schema.Enum {
			if enumVal == strVal {
				found = true
				break
			}
		}
		if !found {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   propName,
				Rule:    "enum_mismatch",
				Message: fmt.Sprintf("property '%s' must be one of: %v", propName, schema.Enum),
				Value:   strVal,
			})
		}
	}

	return nil
}

func (av *AdvancedValidator) validateIntProperty(propName string, value interface{}, schema PropertySchema, result *AdvancedValidationResult) error {
	var intVal int64

	switch v := value.(type) {
	case int:
		intVal = int64(v)
	case int64:
		intVal = v
	case float64:
		intVal = int64(v)
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   propName,
			Rule:    "type_mismatch",
			Message: fmt.Sprintf("property '%s' must be an integer", propName),
			Value:   reflect.TypeOf(value).String(),
		})
		return fmt.Errorf("type mismatch")
	}

	// Validar rango
	if schema.Min != nil {
		var minVal int64
		switch v := schema.Min.(type) {
		case int:
			minVal = int64(v)
		case int64:
			minVal = v
		case float64:
			minVal = int64(v)
		default:
			// Si no puede convertir, saltar validación
			goto validateMax
		}

		if intVal < minVal {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   propName,
				Rule:    "min_value",
				Message: fmt.Sprintf("property '%s' must be at least %v", propName, schema.Min),
				Value:   intVal,
			})
		}
	}

validateMax:
	if schema.Max != nil {
		var maxVal int64
		switch v := schema.Max.(type) {
		case int:
			maxVal = int64(v)
		case int64:
			maxVal = v
		case float64:
			maxVal = int64(v)
		default:
			// Si no puede convertir, saltar validación
			return nil
		}

		if intVal > maxVal {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   propName,
				Rule:    "max_value",
				Message: fmt.Sprintf("property '%s' must be at most %v", propName, schema.Max),
				Value:   intVal,
			})
		}
	}

	return nil
}

func (av *AdvancedValidator) validateNumberProperty(propName string, value interface{}, schema PropertySchema, result *AdvancedValidationResult) error {
	var floatVal float64

	switch v := value.(type) {
	case float64:
		floatVal = v
	case int:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   propName,
			Rule:    "type_mismatch",
			Message: fmt.Sprintf("property '%s' must be a number", propName),
			Value:   reflect.TypeOf(value).String(),
		})
		return fmt.Errorf("type mismatch")
	}

	// Validar rango
	if schema.Min != nil {
		if minVal, ok := schema.Min.(float64); ok && floatVal < minVal {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   propName,
				Rule:    "min_value",
				Message: fmt.Sprintf("property '%s' must be at least %v", propName, schema.Min),
				Value:   floatVal,
			})
		}
	}

	if schema.Max != nil {
		if maxVal, ok := schema.Max.(float64); ok && floatVal > maxVal {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   propName,
				Rule:    "max_value",
				Message: fmt.Sprintf("property '%s' must be at most %v", propName, schema.Max),
				Value:   floatVal,
			})
		}
	}

	return nil
}

func (av *AdvancedValidator) validateBoolProperty(propName string, value interface{}, schema PropertySchema, result *AdvancedValidationResult) error {
	_, ok := value.(bool)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   propName,
			Rule:    "type_mismatch",
			Message: fmt.Sprintf("property '%s' must be a boolean", propName),
			Value:   reflect.TypeOf(value).String(),
		})
		return fmt.Errorf("type mismatch")
	}

	return nil
}
