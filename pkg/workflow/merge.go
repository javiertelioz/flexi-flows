package workflow

import (
	"context"
	"fmt"
	"reflect"
)

// MergeStrategy define cómo se deben fusionar los datos
type MergeStrategy string

const (
	MergeDeep     MergeStrategy = "deep"     // Fusión profunda recursiva
	MergeShallow  MergeStrategy = "shallow"  // Fusión superficial
	MergeOverride MergeStrategy = "override" // Sobrescribir valores
	MergeAppend   MergeStrategy = "append"   // Agregar a arrays
)

// MergeNode representa un nodo que fusiona múltiples fuentes de datos
type MergeNode struct {
	Node[interface{}]
	Sources     []string      // Campos fuente a fusionar
	Strategy    MergeStrategy // Estrategia de fusión
	CustomRules map[string]interface{} // Reglas personalizadas por campo
	IgnoreEmpty bool         // Si ignorar valores vacíos
}

// Execute fusiona datos de múltiples fuentes
func (mn *MergeNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(mn.ID, mn.Type, "context cancelled before merge", ctx.Err())
	default:
	}

	// Convertir datos a map para facilitar fusión
	dataMap, err := mn.convertToMap(data)
	if err != nil {
		return nil, NewWorkflowError(mn.ID, mn.Type, "failed to convert data for merge", err)
	}

	result := make(map[string]interface{})

	// Si no se especifican fuentes, usar todos los campos
	sources := mn.Sources
	if len(sources) == 0 {
		for key := range dataMap {
			sources = append(sources, key)
		}
	}

	// Fusionar cada fuente
	for _, source := range sources {
		select {
		case <-ctx.Done():
			return nil, NewWorkflowError(mn.ID, mn.Type, "context cancelled during merge", ctx.Err())
		default:
		}

		if value, exists := dataMap[source]; exists {
			if err := mn.mergeValue(result, source, value); err != nil {
				return nil, NewWorkflowError(mn.ID, mn.Type,
					fmt.Sprintf("failed to merge field %s", source), err)
			}
		}
	}

	return result, nil
}

// mergeValue fusiona un valor específico en el resultado
func (mn *MergeNode) mergeValue(result map[string]interface{}, key string, value interface{}) error {
	// Verificar si se debe ignorar valores vacíos
	if mn.IgnoreEmpty && mn.isEmpty(value) {
		return nil
	}

	// Aplicar regla personalizada si existe
	if customRule, exists := mn.CustomRules[key]; exists {
		return mn.applyCustomRule(result, key, value, customRule)
	}

	// Aplicar estrategia general
	switch mn.Strategy {
	case MergeDeep:
		return mn.mergeDeep(result, key, value)
	case MergeShallow:
		return mn.mergeShallow(result, key, value)
	case MergeOverride:
		result[key] = value
		return nil
	case MergeAppend:
		return mn.mergeAppend(result, key, value)
	default:
		// Estrategia por defecto: override
		result[key] = value
		return nil
	}
}

// mergeDeep realiza fusión profunda recursiva
func (mn *MergeNode) mergeDeep(result map[string]interface{}, key string, value interface{}) error {
	existing, exists := result[key]
	if !exists {
		result[key] = value
		return nil
	}

	// Si ambos son maps, fusionar recursivamente
	if existingMap, ok := existing.(map[string]interface{}); ok {
		if valueMap, ok := value.(map[string]interface{}); ok {
			mergedMap := make(map[string]interface{})

			// Copiar valores existentes
			for k, v := range existingMap {
				mergedMap[k] = v
			}

			// Fusionar valores nuevos
			for k, v := range valueMap {
				if existingVal, exists := mergedMap[k]; exists {
					if existingSubMap, ok := existingVal.(map[string]interface{}); ok {
						if valueSubMap, ok := v.(map[string]interface{}); ok {
							subResult := make(map[string]interface{})
							for subK, subV := range existingSubMap {
								subResult[subK] = subV
							}
							if err := mn.mergeDeep(subResult, k, valueSubMap); err != nil {
								return err
							}
							mergedMap[k] = subResult[k]
							continue
						}
					}
				}
				mergedMap[k] = v
			}

			result[key] = mergedMap
			return nil
		}
	}

	// Si ambos son arrays, concatenar
	if existingSlice, ok := existing.([]interface{}); ok {
		if valueSlice, ok := value.([]interface{}); ok {
			result[key] = append(existingSlice, valueSlice...)
			return nil
		}
	}

	// En otros casos, sobrescribir
	result[key] = value
	return nil
}

// mergeShallow realiza fusión superficial
func (mn *MergeNode) mergeShallow(result map[string]interface{}, key string, value interface{}) error {
	existing, exists := result[key]
	if !exists {
		result[key] = value
		return nil
	}

	// Solo fusionar si ambos son maps del mismo nivel
	if existingMap, ok := existing.(map[string]interface{}); ok {
		if valueMap, ok := value.(map[string]interface{}); ok {
			mergedMap := make(map[string]interface{})

			// Copiar valores existentes
			for k, v := range existingMap {
				mergedMap[k] = v
			}

			// Agregar valores nuevos (sobrescribir si existen)
			for k, v := range valueMap {
				mergedMap[k] = v
			}

			result[key] = mergedMap
			return nil
		}
	}

	// En otros casos, sobrescribir
	result[key] = value
	return nil
}

// mergeAppend agrega valores a arrays
func (mn *MergeNode) mergeAppend(result map[string]interface{}, key string, value interface{}) error {
	existing, exists := result[key]
	if !exists {
		// Si no existe, crear array con el valor
		result[key] = []interface{}{value}
		return nil
	}

	// Si existe y es array, agregar
	if existingSlice, ok := existing.([]interface{}); ok {
		if valueSlice, ok := value.([]interface{}); ok {
			result[key] = append(existingSlice, valueSlice...)
		} else {
			result[key] = append(existingSlice, value)
		}
		return nil
	}

	// Si existe pero no es array, convertir a array
	result[key] = []interface{}{existing, value}
	return nil
}

// applyCustomRule aplica una regla personalizada
func (mn *MergeNode) applyCustomRule(result map[string]interface{}, key string, value interface{}, rule interface{}) error {
	ruleMap, ok := rule.(map[string]interface{})
	if !ok {
		return fmt.Errorf("custom rule for field %s must be a map", key)
	}

	strategyStr, ok := ruleMap["strategy"].(string)
	if !ok {
		return fmt.Errorf("custom rule for field %s must specify strategy", key)
	}

	strategy := MergeStrategy(strategyStr)
	originalStrategy := mn.Strategy
	mn.Strategy = strategy

	err := mn.mergeValue(result, key, value)

	mn.Strategy = originalStrategy
	return err
}

// isEmpty verifica si un valor está vacío
func (mn *MergeNode) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}

// convertToMap convierte los datos de entrada a un map
func (mn *MergeNode) convertToMap(data interface{}) (map[string]interface{}, error) {
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
