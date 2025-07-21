package workflow

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// FilterCondition define una condición de filtrado
type FilterCondition struct {
	Field         string      `json:"field"`
	Operator      string      `json:"operator"` // eq, ne, gt, lt, gte, lte, contains, regex, in, not_in
	Value         interface{} `json:"value"`
	CaseSensitive bool        `json:"case_sensitive,omitempty"`
}

// FilterNode representa un nodo que filtra datos según condiciones
type FilterNode struct {
	Node[interface{}]
	Conditions   []FilterCondition
	Logic        string      // "and" o "or" para combinar condiciones
	CustomFilter interface{} // Función personalizada de filtrado
	KeepMatched  bool        // Si mantener elementos que coinciden (true) o que no coinciden (false)
}

// FilterResult contiene el resultado de la operación de filtrado
type FilterResult struct {
	Items    []interface{}          `json:"items"`
	Count    int                    `json:"count"`
	Original []interface{}          `json:"original,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Execute filtra los datos según los criterios definidos
func (fn *FilterNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(fn.ID, fn.Type, "context cancelled before filter", ctx.Err())
	default:
	}

	// Convertir datos a slice si es necesario
	items, err := fn.convertToSlice(data)
	if err != nil {
		return nil, NewWorkflowError(fn.ID, fn.Type, "failed to convert data to filterable format", err)
	}

	var filtered []interface{}
	metadata := make(map[string]interface{})

	// Si hay función personalizada, usarla
	if fn.CustomFilter != nil {
		filtered, err = fn.applyCustomFilter(ctx, items)
		if err != nil {
			return nil, NewWorkflowError(fn.ID, fn.Type, "custom filter failed", err)
		}
	} else {
		// Aplicar filtros con condiciones
		for _, item := range items {
			// Verificar cancelación en cada iteración
			select {
			case <-ctx.Done():
				return nil, NewWorkflowError(fn.ID, fn.Type, "context cancelled during filtering", ctx.Err())
			default:
			}

			matches := fn.evaluateConditions(item)
			if (matches && fn.KeepMatched) || (!matches && !fn.KeepMatched) {
				filtered = append(filtered, item)
			}
		}
	}

	// Preparar metadata
	metadata["original_count"] = len(items)
	metadata["filtered_count"] = len(filtered)
	metadata["conditions_count"] = len(fn.Conditions)
	metadata["logic"] = fn.Logic
	metadata["keep_matched"] = fn.KeepMatched

	result := &FilterResult{
		Items:    filtered,
		Count:    len(filtered),
		Metadata: metadata,
	}

	// Incluir datos originales si se requiere
	if fn.shouldKeepOriginal() {
		result.Original = items
	}

	return result, nil
}

// applyCustomFilter aplica una función personalizada de filtrado
func (fn *FilterNode) applyCustomFilter(ctx context.Context, items []interface{}) ([]interface{}, error) {
	funcValue := reflect.ValueOf(fn.CustomFilter)
	funcType := funcValue.Type()

	if funcType.Kind() != reflect.Func {
		return nil, fmt.Errorf("custom filter is not a function")
	}

	var filtered []interface{}

	for _, item := range items {
		// Verificar cancelación
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during custom filter")
		default:
		}

		// Preparar argumentos para la función
		var args []reflect.Value

		// Verificar si la función espera context como primer parámetro
		if funcType.NumIn() > 0 {
			firstParamType := funcType.In(0)
			if firstParamType.String() == "context.Context" {
				args = append(args, reflect.ValueOf(ctx))
				if funcType.NumIn() > 1 {
					args = append(args, reflect.ValueOf(item))
				}
			} else {
				args = append(args, reflect.ValueOf(item))
			}
		}

		// Verificar que el número de argumentos sea correcto
		if len(args) != funcType.NumIn() {
			return nil, fmt.Errorf("custom filter function expects %d arguments, got %d", funcType.NumIn(), len(args))
		}

		// Ejecutar la función
		results := funcValue.Call(args)

		// Procesar resultado
		if len(results) == 1 {
			// Solo un resultado boolean
			if results[0].Type().Kind() == reflect.Bool {
				if results[0].Bool() == fn.KeepMatched {
					filtered = append(filtered, item)
				}
			}
		} else if len(results) == 2 {
			// Resultado boolean y error
			if !results[1].IsNil() {
				return nil, results[1].Interface().(error)
			}
			if results[0].Type().Kind() == reflect.Bool {
				if results[0].Bool() == fn.KeepMatched {
					filtered = append(filtered, item)
				}
			}
		}
	}

	return filtered, nil
}

// evaluateConditions evalúa todas las condiciones para un item
func (fn *FilterNode) evaluateConditions(item interface{}) bool {
	if len(fn.Conditions) == 0 {
		return true
	}

	itemMap, err := fn.convertToMap(item)
	if err != nil {
		return false
	}

	results := make([]bool, len(fn.Conditions))

	// Evaluar cada condición
	for i, condition := range fn.Conditions {
		results[i] = fn.evaluateCondition(itemMap, condition)
	}

	// Combinar resultados según lógica
	if strings.ToLower(fn.Logic) == "or" {
		// OR: al menos una condición debe ser verdadera
		for _, result := range results {
			if result {
				return true
			}
		}
		return false
	} else {
		// AND (por defecto): todas las condiciones deben ser verdaderas
		for _, result := range results {
			if !result {
				return false
			}
		}
		return true
	}
}

// evaluateCondition evalúa una condición específica
func (fn *FilterNode) evaluateCondition(itemMap map[string]interface{}, condition FilterCondition) bool {
	fieldValue, exists := itemMap[condition.Field]
	if !exists {
		return condition.Operator == "ne" // Si el campo no existe, solo coincide con "not equals"
	}

	switch condition.Operator {
	case "eq", "equals":
		return fn.valuesEqual(fieldValue, condition.Value, condition.CaseSensitive)
	case "ne", "not_equals":
		return !fn.valuesEqual(fieldValue, condition.Value, condition.CaseSensitive)
	case "gt", "greater_than":
		return fn.compareValues(fieldValue, condition.Value) > 0
	case "gte", "greater_than_or_equal":
		return fn.compareValues(fieldValue, condition.Value) >= 0
	case "lt", "less_than":
		return fn.compareValues(fieldValue, condition.Value) < 0
	case "lte", "less_than_or_equal":
		return fn.compareValues(fieldValue, condition.Value) <= 0
	case "contains":
		return fn.stringContains(fieldValue, condition.Value, condition.CaseSensitive)
	case "not_contains":
		return !fn.stringContains(fieldValue, condition.Value, condition.CaseSensitive)
	case "starts_with":
		return fn.stringStartsWith(fieldValue, condition.Value, condition.CaseSensitive)
	case "ends_with":
		return fn.stringEndsWith(fieldValue, condition.Value, condition.CaseSensitive)
	case "regex":
		return fn.regexMatches(fieldValue, condition.Value)
	case "in":
		return fn.valueInList(fieldValue, condition.Value)
	case "not_in":
		return !fn.valueInList(fieldValue, condition.Value)
	case "is_null":
		return fieldValue == nil
	case "is_not_null":
		return fieldValue != nil
	case "is_empty":
		return fn.isEmpty(fieldValue)
	case "is_not_empty":
		return !fn.isEmpty(fieldValue)
	default:
		return false
	}
}

// valuesEqual compara dos valores
func (fn *FilterNode) valuesEqual(a, b interface{}, caseSensitive bool) bool {
	if !caseSensitive {
		if strA, okA := a.(string); okA {
			if strB, okB := b.(string); okB {
				return strings.EqualFold(strA, strB)
			}
		}
	}
	return reflect.DeepEqual(a, b)
}

// compareValues compara dos valores numéricamente
func (fn *FilterNode) compareValues(a, b interface{}) int {
	numA := fn.toFloat64(a)
	numB := fn.toFloat64(b)

	if numA < numB {
		return -1
	} else if numA > numB {
		return 1
	}
	return 0
}

// toFloat64 convierte un valor a float64
func (fn *FilterNode) toFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num
		}
	}
	return 0
}

// stringContains verifica si una string contiene otra
func (fn *FilterNode) stringContains(haystack, needle interface{}, caseSensitive bool) bool {
	haystackStr, ok1 := haystack.(string)
	needleStr, ok2 := needle.(string)
	if !ok1 || !ok2 {
		return false
	}

	if !caseSensitive {
		haystackStr = strings.ToLower(haystackStr)
		needleStr = strings.ToLower(needleStr)
	}

	return strings.Contains(haystackStr, needleStr)
}

// stringStartsWith verifica si una string comienza con otra
func (fn *FilterNode) stringStartsWith(text, prefix interface{}, caseSensitive bool) bool {
	textStr, ok1 := text.(string)
	prefixStr, ok2 := prefix.(string)
	if !ok1 || !ok2 {
		return false
	}

	if !caseSensitive {
		textStr = strings.ToLower(textStr)
		prefixStr = strings.ToLower(prefixStr)
	}

	return strings.HasPrefix(textStr, prefixStr)
}

// stringEndsWith verifica si una string termina con otra
func (fn *FilterNode) stringEndsWith(text, suffix interface{}, caseSensitive bool) bool {
	textStr, ok1 := text.(string)
	suffixStr, ok2 := suffix.(string)
	if !ok1 || !ok2 {
		return false
	}

	if !caseSensitive {
		textStr = strings.ToLower(textStr)
		suffixStr = strings.ToLower(suffixStr)
	}

	return strings.HasSuffix(textStr, suffixStr)
}

// regexMatches verifica si un valor coincide con un patrón regex
func (fn *FilterNode) regexMatches(value, pattern interface{}) bool {
	valueStr, ok1 := value.(string)
	patternStr, ok2 := pattern.(string)
	if !ok1 || !ok2 {
		return false
	}

	matched, err := regexp.MatchString(patternStr, valueStr)
	return err == nil && matched
}

// valueInList verifica si un valor está en una lista
func (fn *FilterNode) valueInList(value, list interface{}) bool {
	listSlice, ok := list.([]interface{})
	if !ok {
		return false
	}

	for _, item := range listSlice {
		if reflect.DeepEqual(value, item) {
			return true
		}
	}
	return false
}

// isEmpty verifica si un valor está vacío
func (fn *FilterNode) isEmpty(value interface{}) bool {
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

// convertToSlice convierte datos a slice para poder filtrar
func (fn *FilterNode) convertToSlice(data interface{}) ([]interface{}, error) {
	if slice, ok := data.([]interface{}); ok {
		return slice, nil
	}

	// Si es un solo elemento, crear slice con ese elemento
	return []interface{}{data}, nil
}

// convertToMap convierte un item a map para facilitar acceso a campos
func (fn *FilterNode) convertToMap(data interface{}) (map[string]interface{}, error) {
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
				fieldName = strings.Split(jsonTag, ",")[0]
			}
			dataMap[fieldName] = v.Field(i).Interface()
		}
	}

	return dataMap, nil
}

// shouldKeepOriginal determina si se deben mantener los datos originales
func (fn *FilterNode) shouldKeepOriginal() bool {
	// Se pueden agregar más condiciones aquí según necesidades
	return false
}
