package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// TransformRule define una regla de transformación
type TransformRule struct {
	SourceField string      `json:"source_field"`
	TargetField string      `json:"target_field"`
	Operation   string      `json:"operation"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

// TransformNode representa un nodo que transforma datos
type TransformNode struct {
	Node[interface{}]
	Rules        []TransformRule
	Mapping      map[string]interface{}
	CustomFunc   interface{} // Función personalizada de transformación
	KeepOriginal bool        // Si mantener los campos originales
}

// Execute transforma los datos según las reglas definidas
func (tn *TransformNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(tn.ID, tn.Type, "context cancelled before transform", ctx.Err())
	default:
	}

	// Convertir datos a map para facilitar transformación
	dataMap, err := tn.convertToMap(data)
	if err != nil {
		return nil, NewWorkflowError(tn.ID, tn.Type, "failed to convert data for transformation", err)
	}

	result := make(map[string]interface{})

	// Mantener datos originales si se requiere
	if tn.KeepOriginal {
		for k, v := range dataMap {
			result[k] = v
		}
	}

	// Aplicar mapping directo si existe
	if tn.Mapping != nil {
		for targetField, sourceValue := range tn.Mapping {
			if sourceField, ok := sourceValue.(string); ok {
				// Si existe como campo en los datos, usarlo
				if value, exists := dataMap[sourceField]; exists {
					result[targetField] = value
				} else {
					// Si no existe como campo, tratar como valor constante
					result[targetField] = sourceValue
				}
			} else {
				// Para valores no-string (números, booleanos, etc.)
				result[targetField] = sourceValue
			}
		}
	}

	// Aplicar reglas de transformación
	for _, rule := range tn.Rules {
		select {
		case <-ctx.Done():
			return nil, NewWorkflowError(tn.ID, tn.Type, "context cancelled during transformation", ctx.Err())
		default:
		}

		if err := tn.applyTransformRule(dataMap, result, rule); err != nil {
			return nil, NewWorkflowError(tn.ID, tn.Type,
				fmt.Sprintf("failed to apply transform rule for field %s", rule.SourceField), err)
		}
	}

	// Si hay función personalizada, usarla
	if tn.CustomFunc != nil {
		customResult, err := tn.executeCustomFunction(ctx, data)
		if err != nil {
			return nil, NewWorkflowError(tn.ID, tn.Type, "custom transform failed", err)
		}
		// Convertir el resultado personalizado a map si es necesario
		if customMap, ok := customResult.(map[string]interface{}); ok {
			result = customMap
		} else {
			// Si no es un map, intentar convertir
			convertedMap, err := tn.convertToMap(customResult)
			if err != nil {
				return nil, NewWorkflowError(tn.ID, tn.Type, "failed to convert custom result to map", err)
			}
			result = convertedMap
		}
	}

	return result, nil
}

// executeCustomFunction ejecuta una función personalizada de transformación
func (tn *TransformNode) executeCustomFunction(ctx context.Context, data interface{}) (interface{}, error) {
	funcValue := reflect.ValueOf(tn.CustomFunc)
	funcType := funcValue.Type()

	if funcType.Kind() != reflect.Func {
		return nil, fmt.Errorf("custom function is not a function")
	}

	// Preparar argumentos para la función
	var args []reflect.Value

	// Verificar si la función espera context como primer parámetro
	if funcType.NumIn() > 0 {
		firstParamType := funcType.In(0)
		if firstParamType.String() == "context.Context" {
			args = append(args, reflect.ValueOf(ctx))
			if funcType.NumIn() > 1 {
				args = append(args, reflect.ValueOf(data))
			}
		} else {
			args = append(args, reflect.ValueOf(data))
		}
	}

	// Verificar que el número de argumentos sea correcto
	if len(args) != funcType.NumIn() {
		return nil, fmt.Errorf("custom function expects %d arguments, got %d", funcType.NumIn(), len(args))
	}

	// Ejecutar la función
	results := funcValue.Call(args)

	// Procesar los resultados
	if len(results) == 0 {
		return nil, nil
	}

	if len(results) == 1 {
		// Solo un resultado, verificar si es error
		if results[0].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !results[0].IsNil() {
				return nil, results[0].Interface().(error)
			}
			return nil, nil
		}
		return results[0].Interface(), nil
	}

	if len(results) == 2 {
		// Dos resultados: (resultado, error)
		var err error
		if !results[1].IsNil() {
			err = results[1].Interface().(error)
		}

		if err != nil {
			return nil, err
		}

		return results[0].Interface(), nil
	}

	return nil, fmt.Errorf("unsupported return signature with %d results", len(results))
}

// applyTransformRule aplica una regla de transformación específica
func (tn *TransformNode) applyTransformRule(sourceData, result map[string]interface{}, rule TransformRule) error {
	sourceValue, exists := sourceData[rule.SourceField]
	if !exists {
		// Si el campo fuente no existe, saltarlo silenciosamente
		return nil
	}

	var transformedValue interface{}
	var err error

	switch rule.Operation {
	case "copy":
		transformedValue = sourceValue
	case "uppercase":
		transformedValue, err = tn.transformToUppercase(sourceValue)
	case "lowercase":
		transformedValue, err = tn.transformToLowercase(sourceValue)
	case "trim":
		transformedValue, err = tn.transformTrim(sourceValue)
	case "format":
		transformedValue, err = tn.transformFormat(sourceValue, rule.Parameters)
	case "convert":
		transformedValue, err = tn.transformConvert(sourceValue, rule.Parameters)
	case "substring":
		transformedValue, err = tn.transformSubstring(sourceValue, rule.Parameters)
	case "replace":
		transformedValue, err = tn.transformReplace(sourceValue, rule.Parameters)
	case "split":
		transformedValue, err = tn.transformSplit(sourceValue, rule.Parameters)
	case "join":
		transformedValue, err = tn.transformJoin(sourceValue, rule.Parameters)
	case "calculate":
		transformedValue, err = tn.transformCalculate(sourceValue, rule.Parameters)
	default:
		return fmt.Errorf("unsupported transformation operation: %s", rule.Operation)
	}

	if err != nil {
		return err
	}

	result[rule.TargetField] = transformedValue
	return nil
}

// transformToUppercase convierte texto a mayúsculas
func (tn *TransformNode) transformToUppercase(value interface{}) (interface{}, error) {
	if str, ok := value.(string); ok {
		return strings.ToUpper(str), nil
	}
	return nil, fmt.Errorf("uppercase operation requires string value")
}

// transformToLowercase convierte texto a minúsculas
func (tn *TransformNode) transformToLowercase(value interface{}) (interface{}, error) {
	if str, ok := value.(string); ok {
		return strings.ToLower(str), nil
	}
	return nil, fmt.Errorf("lowercase operation requires string value")
}

// transformTrim elimina espacios en blanco
func (tn *TransformNode) transformTrim(value interface{}) (interface{}, error) {
	if str, ok := value.(string); ok {
		return strings.TrimSpace(str), nil
	}
	return nil, fmt.Errorf("trim operation requires string value")
}

// transformFormat formatea un valor según un patrón
func (tn *TransformNode) transformFormat(value interface{}, params interface{}) (interface{}, error) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("format operation requires parameters")
	}

	pattern, ok := paramsMap["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("format operation requires pattern parameter")
	}

	return fmt.Sprintf(pattern, value), nil
}

// transformConvert convierte un valor a otro tipo
func (tn *TransformNode) transformConvert(value interface{}, params interface{}) (interface{}, error) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("convert operation requires parameters")
	}

	targetType, ok := paramsMap["type"].(string)
	if !ok {
		return nil, fmt.Errorf("convert operation requires type parameter")
	}

	switch targetType {
	case "string":
		return fmt.Sprintf("%v", value), nil
	case "int":
		if str, ok := value.(string); ok {
			return strconv.Atoi(str)
		}
		if num, ok := value.(float64); ok {
			return int(num), nil
		}
		return nil, fmt.Errorf("cannot convert %T to int", value)
	case "float":
		if str, ok := value.(string); ok {
			return strconv.ParseFloat(str, 64)
		}
		if num, ok := value.(int); ok {
			return float64(num), nil
		}
		return nil, fmt.Errorf("cannot convert %T to float", value)
	case "bool":
		if str, ok := value.(string); ok {
			return strconv.ParseBool(str)
		}
		return nil, fmt.Errorf("cannot convert %T to bool", value)
	default:
		return nil, fmt.Errorf("unsupported conversion type: %s", targetType)
	}
}

// transformSubstring extrae una subcadena
func (tn *TransformNode) transformSubstring(value interface{}, params interface{}) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("substring operation requires string value")
	}

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("substring operation requires parameters")
	}

	start, ok := paramsMap["start"].(float64)
	if !ok {
		return nil, fmt.Errorf("substring operation requires start parameter")
	}

	startIdx := int(start)
	if startIdx < 0 || startIdx >= len(str) {
		return "", nil
	}

	if length, exists := paramsMap["length"]; exists {
		if lengthVal, ok := length.(float64); ok {
			endIdx := startIdx + int(lengthVal)
			if endIdx > len(str) {
				endIdx = len(str)
			}
			return str[startIdx:endIdx], nil
		}
	}

	return str[startIdx:], nil
}

// transformReplace reemplaza texto
func (tn *TransformNode) transformReplace(value interface{}, params interface{}) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("replace operation requires string value")
	}

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("replace operation requires parameters")
	}

	old, ok := paramsMap["old"].(string)
	if !ok {
		return nil, fmt.Errorf("replace operation requires old parameter")
	}

	new, ok := paramsMap["new"].(string)
	if !ok {
		return nil, fmt.Errorf("replace operation requires new parameter")
	}

	return strings.ReplaceAll(str, old, new), nil
}

// transformSplit divide una cadena
func (tn *TransformNode) transformSplit(value interface{}, params interface{}) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("split operation requires string value")
	}

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("split operation requires parameters")
	}

	separator, ok := paramsMap["separator"].(string)
	if !ok {
		return nil, fmt.Errorf("split operation requires separator parameter")
	}

	return strings.Split(str, separator), nil
}

// transformJoin une elementos en una cadena
func (tn *TransformNode) transformJoin(value interface{}, params interface{}) (interface{}, error) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("join operation requires parameters")
	}

	separator, ok := paramsMap["separator"].(string)
	if !ok {
		return nil, fmt.Errorf("join operation requires separator parameter")
	}

	// Convertir value a slice de strings
	if slice, ok := value.([]interface{}); ok {
		strSlice := make([]string, len(slice))
		for i, item := range slice {
			strSlice[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(strSlice, separator), nil
	}

	return nil, fmt.Errorf("join operation requires array value")
}

// transformCalculate realiza cálculos matemáticos básicos
func (tn *TransformNode) transformCalculate(value interface{}, params interface{}) (interface{}, error) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("calculate operation requires parameters")
	}

	operation, ok := paramsMap["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("calculate operation requires operation parameter")
	}

	operand, ok := paramsMap["operand"].(float64)
	if !ok {
		return nil, fmt.Errorf("calculate operation requires operand parameter")
	}

	// Convertir value a número
	var num float64
	switch v := value.(type) {
	case float64:
		num = v
	case int:
		num = float64(v)
	case string:
		var err error
		num, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %s to number", v)
		}
	default:
		return nil, fmt.Errorf("calculate operation requires numeric value")
	}

	switch operation {
	case "add":
		return num + operand, nil
	case "subtract":
		return num - operand, nil
	case "multiply":
		return num * operand, nil
	case "divide":
		if operand == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return num / operand, nil
	default:
		return nil, fmt.Errorf("unsupported calculation operation: %s", operation)
	}
}

// convertToMap convierte los datos de entrada a un map para facilitar el acceso
func (tn *TransformNode) convertToMap(data interface{}) (map[string]interface{}, error) {
	if dataMap, ok := data.(map[string]interface{}); ok {
		return dataMap, nil
	}

	// Intentar deserializar JSON si es string
	if str, ok := data.(string); ok {
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(str), &jsonData); err == nil {
			return jsonData, nil
		}
	}

	// Usar reflection para convertir struct a map
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("data must be a map, JSON string, or struct")
	}

	dataMap := make(map[string]interface{})
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			fieldName := strings.ToLower(field.Name)
			dataMap[fieldName] = v.Field(i).Interface()
		}
	}

	return dataMap, nil
}
