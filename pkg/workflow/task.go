package workflow

import (
	"context"
	"fmt"
	"reflect"
)

// TaskNode representa un nodo que ejecuta una tarea específica
type TaskNode struct {
	Node[interface{}]
	Name string
	Task interface{}
}

// Execute ejecuta la tarea del nodo
func (tn *TaskNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	if tn.Task == nil {
		return nil, NewWorkflowError(tn.GetID(), tn.GetType(), "task is nil", fmt.Errorf("task is nil for node %s", tn.GetID()))
	}

	// Usar reflexión para llamar la función de tarea
	taskValue := reflect.ValueOf(tn.Task)
	taskType := taskValue.Type()

	if taskType.Kind() != reflect.Func {
		return nil, NewWorkflowError(tn.GetID(), tn.GetType(), "task is not a function", fmt.Errorf("task for node %s is not a function", tn.GetID()))
	}

	// Preparar argumentos para la función
	var args []reflect.Value

	// Verificar si la función espera context como primer parámetro
	if taskType.NumIn() > 0 {
		firstParamType := taskType.In(0)
		if firstParamType.String() == "context.Context" {
			args = append(args, reflect.ValueOf(ctx))
			if taskType.NumIn() > 1 {
				args = append(args, reflect.ValueOf(data))
			}
		} else {
			args = append(args, reflect.ValueOf(data))
		}
	}

	// Verificar que el número de argumentos sea correcto
	if len(args) != taskType.NumIn() {
		return nil, NewWorkflowError(tn.GetID(), tn.GetType(), "incorrect number of arguments",
			fmt.Errorf("task for node %s expects %d arguments, got %d", tn.GetID(), taskType.NumIn(), len(args)))
	}

	// Ejecutar la función
	results := taskValue.Call(args)

	// Procesar los resultados
	if len(results) == 0 {
		return nil, nil
	}

	if len(results) == 1 {
		// Solo un resultado, verificar si es error
		if results[0].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !results[0].IsNil() {
				return nil, NewWorkflowError(tn.GetID(), tn.GetType(), "task execution failed",
					results[0].Interface().(error))
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
			return nil, NewWorkflowError(tn.GetID(), tn.GetType(), "task execution failed", err)
		}

		return results[0].Interface(), nil
	}

	return nil, NewWorkflowError(tn.GetID(), tn.GetType(), "unsupported return signature",
		fmt.Errorf("task for node %s has unsupported return signature with %d results", tn.GetID(), len(results)))
}

// GetName devuelve el nombre del nodo
func (tn *TaskNode) GetName() string {
	return tn.Name
}
