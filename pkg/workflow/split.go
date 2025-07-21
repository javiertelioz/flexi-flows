package workflow

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// SplitStrategy define cómo se deben dividir los datos
type SplitStrategy string

const (
	SplitByField     SplitStrategy = "field"     // Dividir por campos específicos
	SplitByPattern   SplitStrategy = "pattern"   // Dividir por patrón
	SplitBySize      SplitStrategy = "size"      // Dividir por tamaño de chunks
	SplitByCondition SplitStrategy = "condition" // Dividir por condición
)

// SplitNode representa un nodo que divide datos en múltiples partes
type SplitNode struct {
	Node[interface{}]
	Strategy     SplitStrategy
	Fields       []string               // Campos a dividir (para strategy field)
	Pattern      string                 // Patrón de división (para strategy pattern)
	ChunkSize    int                    // Tamaño de chunks (para strategy size)
	Condition    interface{}            // Función de condición (para strategy condition)
	Parameters   map[string]interface{} // Parámetros adicionales
	KeepOriginal bool                   // Si mantener los datos originales
}

// SplitResult representa el resultado de una operación de división
type SplitResult struct {
	Parts    []interface{}          `json:"parts"`
	Count    int                    `json:"count"`
	Strategy string                 `json:"strategy"`
	Metadata map[string]interface{} `json:"metadata"`
	Original interface{}            `json:"original,omitempty"`
}

// Execute divide los datos según la estrategia especificada
func (sn *SplitNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(sn.ID, sn.Type, "context cancelled before split", ctx.Err())
	default:
	}

	var parts []interface{}
	var metadata = make(map[string]interface{})
	var err error

	// Validar que hay datos para dividir
	if data == nil {
		return nil, NewWorkflowError(sn.ID, sn.Type, "no data provided for split", fmt.Errorf("data is nil"))
	}

	// Aplicar estrategia de división
	switch sn.Strategy {
	case SplitByField:
		parts, err = sn.splitByFields(data, metadata)
	case SplitByPattern:
		parts, err = sn.splitByPattern(data, metadata)
	case SplitBySize:
		parts, err = sn.splitBySize(data, metadata)
	case SplitByCondition:
		parts, err = sn.splitByCondition(ctx, data, metadata)
	default:
		return nil, NewWorkflowError(sn.ID, sn.Type, fmt.Sprintf("unsupported split strategy: %s", sn.Strategy), fmt.Errorf("invalid strategy"))
	}

	if err != nil {
		return nil, NewWorkflowError(sn.ID, sn.Type, "split operation failed", err)
	}

	result := &SplitResult{
		Parts:    parts,
		Count:    len(parts),
		Strategy: string(sn.Strategy),
		Metadata: metadata,
	}

	if sn.KeepOriginal {
		result.Original = data
	}

	return result, nil
}

// splitByFields divide los datos por campos específicos
func (sn *SplitNode) splitByFields(data interface{}, metadata map[string]interface{}) ([]interface{}, error) {
	dataMap, err := sn.convertToMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert data to map: %w", err)
	}

	if len(sn.Fields) == 0 {
		return nil, fmt.Errorf("no fields specified for field split")
	}

	parts := make([]interface{}, 0, len(sn.Fields))
	metadata["fields"] = sn.Fields

	for _, field := range sn.Fields {
		if value, exists := dataMap[field]; exists {
			part := map[string]interface{}{
				"field": field,
				"value": value,
			}
			parts = append(parts, part)
		}
	}

	return parts, nil
}

// splitByPattern divide los datos usando un patrón
func (sn *SplitNode) splitByPattern(data interface{}, metadata map[string]interface{}) ([]interface{}, error) {
	if sn.Pattern == "" {
		return nil, fmt.Errorf("no pattern specified for pattern split")
	}

	metadata["pattern"] = sn.Pattern

	// Si es string, dividir por el patrón
	if str, ok := data.(string); ok {
		parts := strings.Split(str, sn.Pattern)
		result := make([]interface{}, len(parts))
		for i, part := range parts {
			result[i] = strings.TrimSpace(part)
		}
		return result, nil
	}

	// Si es array, dividir según el patrón como separador numérico
	if slice, ok := data.([]interface{}); ok {
		if sn.Pattern == "," { // Ejemplo: dividir en elementos individuales
			result := make([]interface{}, len(slice))
			copy(result, slice)
			return result, nil
		}
	}

	return nil, fmt.Errorf("pattern split requires string or array data")
}

// splitBySize divide los datos en chunks de tamaño específico
func (sn *SplitNode) splitBySize(data interface{}, metadata map[string]interface{}) ([]interface{}, error) {
	if sn.ChunkSize <= 0 {
		return nil, fmt.Errorf("chunk size must be positive")
	}

	metadata["chunk_size"] = sn.ChunkSize

	// Si es string, dividir en chunks de caracteres
	if str, ok := data.(string); ok {
		parts := make([]interface{}, 0)
		for i := 0; i < len(str); i += sn.ChunkSize {
			end := i + sn.ChunkSize
			if end > len(str) {
				end = len(str)
			}
			parts = append(parts, str[i:end])
		}
		return parts, nil
	}

	// Si es slice, dividir en chunks
	if slice, ok := data.([]interface{}); ok {
		parts := make([]interface{}, 0)
		for i := 0; i < len(slice); i += sn.ChunkSize {
			end := i + sn.ChunkSize
			if end > len(slice) {
				end = len(slice)
			}
			chunk := make([]interface{}, end-i)
			copy(chunk, slice[i:end])
			parts = append(parts, chunk)
		}
		return parts, nil
	}

	return nil, fmt.Errorf("size split requires string or array data")
}

// splitByCondition divide los datos usando una función de condición
func (sn *SplitNode) splitByCondition(ctx context.Context, data interface{}, metadata map[string]interface{}) ([]interface{}, error) {
	if sn.Condition == nil {
		return nil, fmt.Errorf("no condition function specified for condition split")
	}

	// Si es slice, dividir según condición
	if slice, ok := data.([]interface{}); ok {
		var truePart, falsePart []interface{}

		for _, item := range slice {
			// Verificar cancelación en cada iteración
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during condition split")
			default:
			}

			if sn.evaluateCondition(item) {
				truePart = append(truePart, item)
			} else {
				falsePart = append(falsePart, item)
			}
		}

		metadata["true_count"] = len(truePart)
		metadata["false_count"] = len(falsePart)

		return []interface{}{
			map[string]interface{}{
				"condition": true,
				"items":     truePart,
			},
			map[string]interface{}{
				"condition": false,
				"items":     falsePart,
			},
		}, nil
	}

	return nil, fmt.Errorf("condition split requires array data")
}

// evaluateCondition evalúa una condición sobre un item
func (sn *SplitNode) evaluateCondition(item interface{}) bool {
	// Si la condición es una función
	if condFunc, ok := sn.Condition.(func(interface{}) bool); ok {
		return condFunc(item)
	}

	// Si la condición es una función con reflexión
	funcValue := reflect.ValueOf(sn.Condition)
	if funcValue.Kind() == reflect.Func {
		funcType := funcValue.Type()
		if funcType.NumIn() == 1 && funcType.NumOut() == 1 {
			args := []reflect.Value{reflect.ValueOf(item)}
			results := funcValue.Call(args)
			if results[0].Type().Kind() == reflect.Bool {
				return results[0].Bool()
			}
		}
	}

	// Si la condición es un map con criterios
	if condMap, ok := sn.Condition.(map[string]interface{}); ok {
		return sn.evaluateMapCondition(item, condMap)
	}

	// Por defecto, retornar false
	return false
}

// evaluateMapCondition evalúa una condición definida como map
func (sn *SplitNode) evaluateMapCondition(item interface{}, condition map[string]interface{}) bool {
	itemMap, err := sn.convertToMap(item)
	if err != nil {
		return false
	}

	// Evaluar cada criterio en la condición
	for field, expectedValue := range condition {
		if actualValue, exists := itemMap[field]; exists {
			if !sn.valuesEqual(actualValue, expectedValue) {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

// valuesEqual compara dos valores
func (sn *SplitNode) valuesEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// convertToMap convierte los datos de entrada a un map
func (sn *SplitNode) convertToMap(data interface{}) (map[string]interface{}, error) {
	if dataMap, ok := data.(map[string]interface{}); ok {
		return dataMap, nil
	}

	// Usar reflection para convertir struct a map
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("data must be a map or struct")
	}

	dataMap := make(map[string]interface{})
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			fieldName := field.Name
			// Usar tag json si existe
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				fieldName = jsonTag
			}
			dataMap[fieldName] = v.Field(i).Interface()
		}
	}

	return dataMap, nil
}
