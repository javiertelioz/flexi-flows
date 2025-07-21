package workflow

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ValidationNode representa un nodo que valida datos según reglas definidas
type ValidationNode struct {
	Node[interface{}]
	Rules            []ValidationRule
	StopOnFirstError bool
}

// ValidationResult contiene el resultado de la validación
type ValidationResult struct {
	Valid   bool                   `json:"valid"`
	Errors  []ValidationError      `json:"errors,omitempty"`
	Data    interface{}            `json:"data"`
	Summary map[string]interface{} `json:"summary"`
}

// ValidationError representa un error de validación específico
type ValidationError struct {
	Field   string      `json:"field"`
	Rule    string      `json:"rule"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// Execute valida los datos según las reglas definidas
func (vn *ValidationNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(vn.ID, vn.Type, "context cancelled before validation", ctx.Err())
	default:
	}

	result := &ValidationResult{
		Valid:   true,
		Errors:  []ValidationError{},
		Data:    data,
		Summary: make(map[string]interface{}),
	}

	// Convertir datos a map para facilitar acceso
	dataMap, err := vn.convertToMap(data)
	if err != nil {
		return nil, NewWorkflowError(vn.ID, vn.Type, "failed to convert data for validation", err)
	}

	// Aplicar cada regla de validación
	for _, rule := range vn.Rules {
		select {
		case <-ctx.Done():
			return nil, NewWorkflowError(vn.ID, vn.Type, "context cancelled during validation", ctx.Err())
		default:
		}

		if err := vn.validateField(dataMap, rule, result); err != nil {
			if vn.StopOnFirstError {
				result.Valid = false
				return result, NewWorkflowError(vn.ID, vn.Type, "validation failed", err)
			}
		}
	}

	// Preparar resumen
	result.Summary["total_rules"] = len(vn.Rules)
	result.Summary["errors_count"] = len(result.Errors)
	result.Summary["fields_validated"] = vn.getValidatedFields()

	// Si hay errores, marcar como inválido
	if len(result.Errors) > 0 {
		result.Valid = false
	}

	// Si la validación falla y se requiere error, retornar error
	if !result.Valid && vn.StopOnFirstError {
		return result, NewWorkflowError(vn.ID, vn.Type,
			fmt.Sprintf("validation failed with %d errors", len(result.Errors)),
			fmt.Errorf("validation errors: %v", result.Errors))
	}

	return result, nil
}

// validateField valida un campo específico según una regla
func (vn *ValidationNode) validateField(dataMap map[string]interface{}, rule ValidationRule, result *ValidationResult) error {
	value, exists := dataMap[rule.Field]

	// Validar campo requerido
	if rule.Required && (!exists || value == nil || value == "") {
		result.Errors = append(result.Errors, ValidationError{
			Field:   rule.Field,
			Rule:    "required",
			Message: vn.getMessage(rule, "field is required"),
			Value:   value,
		})
		return fmt.Errorf("field %s is required", rule.Field)
	}

	// Si el campo no existe y no es requerido, saltarlo
	if !exists || value == nil {
		return nil
	}

	// Validar según el tipo
	switch rule.Type {
	case "string":
		return vn.validateString(value, rule, result)
	case "int", "integer":
		return vn.validateInteger(value, rule, result)
	case "float", "number":
		return vn.validateFloat(value, rule, result)
	case "email":
		return vn.validateEmail(value, rule, result)
	case "url":
		return vn.validateURL(value, rule, result)
	case "bool", "boolean":
		return vn.validateBoolean(value, rule, result)
	default:
		// Tipo no soportado, pero no es un error crítico
		return nil
	}
}

// validateString valida un campo de tipo string
func (vn *ValidationNode) validateString(value interface{}, rule ValidationRule, result *ValidationResult) error {
	str, ok := value.(string)
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Field:   rule.Field,
			Rule:    "type",
			Message: vn.getMessage(rule, "field must be a string"),
			Value:   value,
		})
		return fmt.Errorf("field %s must be a string", rule.Field)
	}

	// Validar longitud mínima
	if rule.MinValue != nil {
		if minLen, ok := rule.MinValue.(float64); ok && len(str) < int(minLen) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   rule.Field,
				Rule:    "min_length",
				Message: vn.getMessage(rule, fmt.Sprintf("field must be at least %d characters", int(minLen))),
				Value:   len(str),
			})
			return fmt.Errorf("field %s is too short", rule.Field)
		}
	}

	// Validar longitud máxima
	if rule.MaxValue != nil {
		if maxLen, ok := rule.MaxValue.(float64); ok && len(str) > int(maxLen) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   rule.Field,
				Rule:    "max_length",
				Message: vn.getMessage(rule, fmt.Sprintf("field must be at most %d characters", int(maxLen))),
				Value:   len(str),
			})
			return fmt.Errorf("field %s is too long", rule.Field)
		}
	}

	// Validar patrón regex
	if rule.Pattern != "" {
		matched, err := regexp.MatchString(rule.Pattern, str)
		if err != nil {
			return fmt.Errorf("invalid regex pattern for field %s", rule.Field)
		}
		if !matched {
			result.Errors = append(result.Errors, ValidationError{
				Field:   rule.Field,
				Rule:    "pattern",
				Message: vn.getMessage(rule, "field does not match required pattern"),
				Value:   str,
			})
			return fmt.Errorf("field %s does not match pattern", rule.Field)
		}
	}

	return nil
}

// validateEmail valida un campo de tipo email
func (vn *ValidationNode) validateEmail(value interface{}, rule ValidationRule, result *ValidationResult) error {
	str, ok := value.(string)
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Field:   rule.Field,
			Rule:    "type",
			Message: vn.getMessage(rule, "field must be a string"),
			Value:   value,
		})
		return fmt.Errorf("field %s must be a string", rule.Field)
	}

	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailPattern, str)
	if !matched {
		result.Errors = append(result.Errors, ValidationError{
			Field:   rule.Field,
			Rule:    "email",
			Message: vn.getMessage(rule, "field must be a valid email address"),
			Value:   str,
		})
		return fmt.Errorf("field %s is not a valid email", rule.Field)
	}

	return nil
}

// validateInteger valida un campo de tipo entero
func (vn *ValidationNode) validateInteger(value interface{}, rule ValidationRule, result *ValidationResult) error {
	var intVal int64
	var err error

	switch v := value.(type) {
	case int:
		intVal = int64(v)
	case int64:
		intVal = v
	case float64:
		intVal = int64(v)
	case string:
		intVal, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   rule.Field,
				Rule:    "type",
				Message: vn.getMessage(rule, "field must be a valid integer"),
				Value:   value,
			})
			return fmt.Errorf("field %s is not a valid integer", rule.Field)
		}
	default:
		result.Errors = append(result.Errors, ValidationError{
			Field:   rule.Field,
			Rule:    "type",
			Message: vn.getMessage(rule, "field must be an integer"),
			Value:   value,
		})
		return fmt.Errorf("field %s must be an integer", rule.Field)
	}

	// Validar valor mínimo
	if rule.MinValue != nil {
		if minVal, ok := rule.MinValue.(float64); ok && intVal < int64(minVal) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   rule.Field,
				Rule:    "min_value",
				Message: vn.getMessage(rule, fmt.Sprintf("field must be at least %d", int64(minVal))),
				Value:   intVal,
			})
			return fmt.Errorf("field %s is below minimum value", rule.Field)
		}
	}

	// Validar valor máximo
	if rule.MaxValue != nil {
		if maxVal, ok := rule.MaxValue.(float64); ok && intVal > int64(maxVal) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   rule.Field,
				Rule:    "max_value",
				Message: vn.getMessage(rule, fmt.Sprintf("field must be at most %d", int64(maxVal))),
				Value:   intVal,
			})
			return fmt.Errorf("field %s is above maximum value", rule.Field)
		}
	}

	return nil
}

// validateFloat valida un campo de tipo float
func (vn *ValidationNode) validateFloat(value interface{}, rule ValidationRule, result *ValidationResult) error {
	var floatVal float64
	var err error

	switch v := value.(type) {
	case float64:
		floatVal = v
	case int:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	case string:
		floatVal, err = strconv.ParseFloat(v, 64)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   rule.Field,
				Rule:    "type",
				Message: vn.getMessage(rule, "field must be a valid number"),
				Value:   value,
			})
			return fmt.Errorf("field %s is not a valid number", rule.Field)
		}
	default:
		result.Errors = append(result.Errors, ValidationError{
			Field:   rule.Field,
			Rule:    "type",
			Message: vn.getMessage(rule, "field must be a number"),
			Value:   value,
		})
		return fmt.Errorf("field %s must be a number", rule.Field)
	}

	// Validar valor mínimo
	if rule.MinValue != nil {
		if minVal, ok := rule.MinValue.(float64); ok && floatVal < minVal {
			result.Errors = append(result.Errors, ValidationError{
				Field:   rule.Field,
				Rule:    "min_value",
				Message: vn.getMessage(rule, fmt.Sprintf("field must be at least %.2f", minVal)),
				Value:   floatVal,
			})
			return fmt.Errorf("field %s is below minimum value", rule.Field)
		}
	}

	// Validar valor máximo
	if rule.MaxValue != nil {
		if maxVal, ok := rule.MaxValue.(float64); ok && floatVal > maxVal {
			result.Errors = append(result.Errors, ValidationError{
				Field:   rule.Field,
				Rule:    "max_value",
				Message: vn.getMessage(rule, fmt.Sprintf("field must be at most %.2f", maxVal)),
				Value:   floatVal,
			})
			return fmt.Errorf("field %s is above maximum value", rule.Field)
		}
	}

	return nil
}

// validateURL valida un campo de tipo URL
func (vn *ValidationNode) validateURL(value interface{}, rule ValidationRule, result *ValidationResult) error {
	str, ok := value.(string)
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Field:   rule.Field,
			Rule:    "type",
			Message: vn.getMessage(rule, "field must be a string"),
			Value:   value,
		})
		return fmt.Errorf("field %s must be a string", rule.Field)
	}

	urlPattern := `^https?://[^\s/$.?#].[^\s]*$`
	matched, _ := regexp.MatchString(urlPattern, str)
	if !matched {
		result.Errors = append(result.Errors, ValidationError{
			Field:   rule.Field,
			Rule:    "url",
			Message: vn.getMessage(rule, "field must be a valid URL"),
			Value:   str,
		})
		return fmt.Errorf("field %s is not a valid URL", rule.Field)
	}

	return nil
}

// validateBoolean valida un campo de tipo boolean
func (vn *ValidationNode) validateBoolean(value interface{}, rule ValidationRule, result *ValidationResult) error {
	switch v := value.(type) {
	case bool:
		return nil
	case string:
		if strings.ToLower(v) == "true" || strings.ToLower(v) == "false" {
			return nil
		}
	}

	result.Errors = append(result.Errors, ValidationError{
		Field:   rule.Field,
		Rule:    "type",
		Message: vn.getMessage(rule, "field must be a boolean"),
		Value:   value,
	})
	return fmt.Errorf("field %s must be a boolean", rule.Field)
}

// convertToMap convierte los datos de entrada a un map para facilitar el acceso
func (vn *ValidationNode) convertToMap(data interface{}) (map[string]interface{}, error) {
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
			fieldName := strings.ToLower(field.Name)
			dataMap[fieldName] = v.Field(i).Interface()
		}
	}

	return dataMap, nil
}

// getMessage obtiene el mensaje personalizado o usa el por defecto
func (vn *ValidationNode) getMessage(rule ValidationRule, defaultMessage string) string {
	if rule.Message != "" {
		return rule.Message
	}
	return defaultMessage
}

// getValidatedFields obtiene la lista de campos que fueron validados
func (vn *ValidationNode) getValidatedFields() []string {
	fields := make([]string, len(vn.Rules))
	for i, rule := range vn.Rules {
		fields[i] = rule.Field
	}
	return fields
}
