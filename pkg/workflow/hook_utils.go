package workflow

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// validateHookFunction valida que una función tenga la firma correcta para ser un hook
func validateHookFunction(fn interface{}) error {
	if fn == nil {
		return fmt.Errorf("hook function cannot be nil")
	}

	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		return fmt.Errorf("hook must be a function")
	}

	fnType := fnValue.Type()

	// Verificar firmas válidas:
	// func() error
	// func(data interface{}) error
	// func(ctx context.Context, hookCtx *HookContext) error

	switch fnType.NumIn() {
	case 0:
		// func() error
		if fnType.NumOut() != 1 || !fnType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return fmt.Errorf("hook function with no parameters must return only error")
		}
	case 1:
		// func(data interface{}) error
		if fnType.NumOut() != 1 || !fnType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return fmt.Errorf("hook function with one parameter must return only error")
		}
	case 2:
		// func(ctx context.Context, hookCtx *HookContext) error
		if fnType.NumOut() != 1 || !fnType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return fmt.Errorf("hook function with two parameters must return only error")
		}

		// Verificar tipos de parámetros
		contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
		hookCtxType := reflect.TypeOf((*HookContext)(nil))

		if !fnType.In(0).Implements(contextType) {
			return fmt.Errorf("first parameter must be context.Context")
		}
		if fnType.In(1) != hookCtxType {
			return fmt.Errorf("second parameter must be *HookContext")
		}
	default:
		return fmt.Errorf("hook function must have 0, 1, or 2 parameters")
	}

	return nil
}

// executeHookFunction ejecuta una función de hook usando reflection
func executeHookFunction(ctx context.Context, fn interface{}, hookCtx *HookContext) error {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	var args []reflect.Value

	switch fnType.NumIn() {
	case 0:
		// func() error
		args = []reflect.Value{}
	case 1:
		// func(data interface{}) error
		args = []reflect.Value{reflect.ValueOf(hookCtx.Data)}
	case 2:
		// func(ctx context.Context, hookCtx *HookContext) error
		args = []reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(hookCtx),
		}
	default:
		return fmt.Errorf("unsupported hook function signature")
	}

	results := fnValue.Call(args)
	if len(results) != 1 {
		return fmt.Errorf("hook function must return exactly one value (error)")
	}

	if results[0].IsNil() {
		return nil
	}

	if err, ok := results[0].Interface().(error); ok {
		return err
	}

	return fmt.Errorf("hook function must return an error")
}

// DefaultLogger es una implementación simple de Logger
type DefaultLogger struct{}

func (dl *DefaultLogger) Log(level LogLevel, message string, fields map[string]interface{}) {
	levelStr := ""
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
	case INFO:
		levelStr = "INFO"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	}

	fmt.Printf("[%s] %s", levelStr, message)
	if len(fields) > 0 {
		fmt.Printf(" - %+v", fields)
	}
	fmt.Println()
}

// DefaultMetricsCollector es una implementación simple de MetricsCollector
type DefaultMetricsCollector struct {
	counters map[string]int64
	mu       sync.RWMutex
}

func NewDefaultMetricsCollector() *DefaultMetricsCollector {
	return &DefaultMetricsCollector{
		counters: make(map[string]int64),
	}
}

func (dmc *DefaultMetricsCollector) RecordHookExecution(hookType HookType, nodeType NodeType, duration time.Duration) {
	// En una implementación real, esto podría enviar métricas a Prometheus, DataDog, etc.
	fmt.Printf("Hook %s on %s took %v\n", hookType, nodeType, duration)
}

func (dmc *DefaultMetricsCollector) RecordNodeExecution(nodeID string, nodeType NodeType, duration time.Duration, success bool) {
	fmt.Printf("Node %s (%s) took %v, success: %t\n", nodeID, nodeType, duration, success)
}

func (dmc *DefaultMetricsCollector) IncrementCounter(name string, tags map[string]string) {
	dmc.mu.Lock()
	defer dmc.mu.Unlock()
	dmc.counters[name]++
}

func (dmc *DefaultMetricsCollector) GetCounter(name string) int64 {
	dmc.mu.RLock()
	defer dmc.mu.RUnlock()
	return dmc.counters[name]
}
