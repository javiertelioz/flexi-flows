# Flexi-Flows: Plugin de Workflows Simplificado

## Descripción General

Flexi-Flows es un plugin de workflows en Go que permite crear flujos de trabajo complejos usando configuración declarativa en JSON o YAML. El sistema ha sido completamente mejorado para ser más simple, robusto y fácil de usar.

## Características Principales

### ✅ Configuración Simplificada
- Soporte completo para JSON y YAML
- Navegación declarativa (no más edges complicados)
- Auto-descubrimiento de funciones
- Validación automática de configuración

### ✅ Sistema de Hooks Avanzado
- Hooks globales y por nodo
- Múltiples tipos de hooks: `before`, `after`, `success`, `error`, `complete`
- Conversión automática de funciones a hooks
- Contexto completo en cada hook

### ✅ Manejo de Errores Robusto
- Errores con contexto completo y stack trace
- Trazabilidad por nodo y tipo
- Metadata adicional para debugging
- Validación de configuración

### ✅ Tipos de Nodos Ampliados
- `task` - Ejecución de tareas básicas
- `conditional` - Lógica if/else
- `decision` - Ramificación condicional (antes Branch)
- `loop` - Iteración sobre colecciones (antes Foreach)
- `parallel` - Ejecución paralela
- `subflow` - Sub-workflows (antes SubDag)
- `http` - Peticiones HTTP
- `delay` - Retrasos temporales
- `transform` - Transformación de datos
- `validation` - Validación de datos
- `merge` - Fusión de datos
- `split` - División de datos
- `filter` - Filtrado de datos

## 🚧 Nodos Pendientes - Implementación Futura

### 🔧 Nodos de Integración
Los siguientes nodos están planificados para futuras versiones para mejorar la integración empresarial:

- **`database`** - Para operaciones CRUD en bases de datos (MySQL, PostgreSQL, MongoDB)
- **`cache`** - Para operaciones de caché (Redis, Memcached)
- **`queue`** - Para integración con colas de mensajes (RabbitMQ, AWS SQS, Apache Kafka)
- **`webhook`** - Para recibir y procesar webhooks
- **`email`** - Para envío de emails (SMTP, SendGrid, AWS SES)
- **`notification`** - Para notificaciones push, Slack, Teams, Discord

### 🎮 Nodos de Control Avanzado
Nodos para patrones de control más sofisticados:

- **`retry`** - Reintentos con backoff exponencial y políticas avanzadas
- **`circuit_breaker`** - Implementación del patrón Circuit Breaker
- **`batch`** - Procesamiento en lotes con ventanas deslizantes
- **`scheduler`** - Programación de ejecuciones (cron-like, intervals)

### 📈 Nodos de Monitoreo
Para observabilidad y monitoreo avanzado:

- **`metrics`** - Recolección de métricas (Prometheus, InfluxDB)
- **`audit`** - Auditoría avanzada con trazabilidad completa
- **`health_check`** - Verificaciones de salud de servicios externos

> **Nota:** Si necesitas alguno de estos nodos para tu proyecto, puedes contribuir con su implementación o abrir un issue para solicitar su priorización.

## Guía de Uso

### Instalación

```go
import "github.com/javiertelioz/flexi-flows/pkg/workflow"
```

### Ejemplo Básico

#### 1. Configuración JSON Simplificada

```json
{
  "name": "mi_workflow",
  "description": "Un workflow de ejemplo",
  "start_node": "tarea_inicial",
  "nodes": [
    {
      "id": "tarea_inicial",
      "type": "task",
      "function": "ProcesarDatos",
      "next": ["validacion"]
    },
    {
      "id": "validacion",
      "type": "validation",
      "rules": [
        {
          "field": "email",
          "type": "email",
          "required": true
        }
      ],
      "on_success": ["enviar_email"],
      "on_error": ["manejar_error"]
    },
    {
      "id": "enviar_email",
      "type": "http",
      "url": "https://api.email.com/send",
      "method": "POST"
    },
    {
      "id": "manejar_error",
      "type": "task",
      "function": "ManejarError"
    }
  ],
  "hooks": {
    "audit_log": {
      "type": "after",
      "function": "LogAuditoria"
    }
  }
}
```

#### 2. Código Go

```go
func main() {
    // Crear el manager de workflows
    wm := workflow.NewWorkflowManager()
    
    // Registrar funciones de tareas
    wm.RegisterTask("ProcesarDatos", func(data map[string]interface{}) (map[string]interface{}, error) {
        data["procesado"] = true
        return data, nil
    })
    
    wm.RegisterTask("ManejarError", func(data map[string]interface{}) (map[string]interface{}, error) {
        data["error_manejado"] = true
        return data, nil
    })
    
    wm.RegisterTask("LogAuditoria", func() error {
        fmt.Println("Nodo ejecutado para auditoría")
        return nil
    })
    
    // Cargar configuración
    parser := config.NewConfigParser()
    cfg, err := parser.ParseFromFile("mi_workflow.json")
    if err != nil {
        log.Fatal(err)
    }
    
    err = wm.LoadFromConfig(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    // Ejecutar workflow
    ctx := context.Background()
    resultado, err := wm.Execute(ctx, map[string]interface{}{
        "email": "usuario@ejemplo.com",
        "datos": "información inicial",
    })
    
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Resultado: %v", resultado)
    }
}
```

### Tipos de Nodos Disponibles

#### Task Node
```yaml
- id: "mi_tarea"
  type: "task"
  function: "MiFuncion"
  timeout: "30s"
  retry_count: 3
```

#### Conditional Node
```yaml
- id: "decision"
  type: "conditional"
  condition: "data.edad >= 18"
  true_path: ["proceso_adulto"]
  false_path: ["proceso_menor"]
```

#### HTTP Node
```yaml
- id: "api_call"
  type: "http"
  url: "https://api.ejemplo.com/data"
  method: "POST"
  headers:
    Content-Type: "application/json"
  body:
    id: "{{data.user_id}}"
```

#### Loop Node
```yaml
- id: "procesar_items"
  type: "loop"
  iterator: "item"
  collection: "data.items"
  next: ["procesar_item"]
```

#### Validation Node
```yaml
- id: "validar_entrada"
  type: "validation"
  rules:
    - field: "email"
      type: "email"
      required: true
    - field: "edad"
      type: "number"
      min: 0
      max: 120
```

### Sistema de Hooks

#### Tipos de Hooks Disponibles
- `before` - Antes de ejecutar un nodo
- `after` - Después de ejecutar un nodo
- `success` - Solo cuando un nodo tiene éxito
- `error` - Solo cuando un nodo falla
- `complete` - Siempre al finalizar (éxito o error)

#### Registro de Hooks

```go
// Hook global para todos los nodos
wm.RegisterHookFunc(workflow.BeforeExecution, func() error {
    fmt.Println("Iniciando ejecución de nodo...")
    return nil
})

// Hook con acceso a datos
wm.RegisterHookFunc(workflow.AfterExecution, func(data interface{}) error {
    fmt.Printf("Nodo completado con datos: %v\n", data)
    return nil
})

// Hook con contexto completo
wm.RegisterHook(workflow.OnError, func(ctx context.Context, hookCtx *workflow.HookContext) error {
    fmt.Printf("Error en nodo %s: %v\n", hookCtx.NodeID, hookCtx.Metadata["error"])
    return nil
})
```

### Manejo de Errores

El sistema proporciona errores detallados con contexto:

```go
if err != nil {
    if workflowErr, ok := err.(*workflow.WorkflowError); ok {
        fmt.Printf("Error en nodo: %s\n", workflowErr.NodeID)
        fmt.Printf("Tipo de nodo: %s\n", workflowErr.NodeType)
        fmt.Printf("Mensaje: %s\n", workflowErr.Message)
        fmt.Printf("Timestamp: %v\n", workflowErr.Timestamp)
        fmt.Printf("Contexto: %v\n", workflowErr.Context)
        fmt.Printf("Stack trace: %v\n", workflowErr.Stack)
    }
}
```

### Navegación Simplificada

En lugar de definir edges por separado, puedes usar navegación directa en los nodos:

```yaml
nodes:
  - id: "inicio"
    type: "task"
    function: "Iniciar"
    next: ["paso2"]              # Navegación normal
    on_success: ["log_exito"]    # Solo si tiene éxito
    on_error: ["manejar_error"]  # Solo si hay error
```

### Variables y Templating

```yaml
variables:
  api_url: "https://api.ejemplo.com"
  timeout: "30s"

nodes:
  - id: "api_call"
    type: "http"
    url: "${api_url}/data"
    timeout: "${timeout}"
```

## Mejores Prácticas

### 1. Organización de Funciones
```go
// Agrupa funciones relacionadas
type UserService struct{}

func (u *UserService) ValidateUser(data map[string]interface{}) (map[string]interface{}, error) {
    // validación
}

func (u *UserService) CreateUser(data map[string]interface{}) (map[string]interface{}, error) {
    // creación
}

// Registra el service
userService := &UserService{}
wm.RegisterTask("ValidateUser", userService.ValidateUser)
wm.RegisterTask("CreateUser", userService.CreateUser)
```

### 2. Manejo de Errores Específicos
```go
wm.RegisterTask("TaskWithSpecificError", func(data map[string]interface{}) (map[string]interface{}, error) {
    if val, ok := data["required_field"]; !ok || val == nil {
        return nil, workflow.NewValidationError("required_field", "", "campo requerido faltante")
    }
    // lógica de la tarea
    return data, nil
})
```

### 3. Configuración por Entorno
```yaml
# config/development.yaml
name: "mi_workflow_dev"
variables:
  api_url: "https://dev-api.ejemplo.com"
  debug: true

# config/production.yaml  
name: "mi_workflow_prod"
variables:
  api_url: "https://api.ejemplo.com"
  debug: false
```

### 4. Testing
```go
func TestMiWorkflow(t *testing.T) {
    wm := workflow.NewWorkflowManager()
    
    // Setup mocks
    wm.RegisterTask("MockTask", func(data map[string]interface{}) (map[string]interface{}, error) {
        data["mock"] = true
        return data, nil
    })
    
    // Load config y ejecutar tests
    // ... resto del test
}
```

## Migración desde la Versión Anterior

### Cambios Principales
1. `Branch` → `DecisionNode`
2. `Foreach` → `LoopNode`  
3. `SubDag` → `SubflowNode`
4. Edges opcionales (usa navegación en nodos)
5. Hooks mejorados y centralizados
6. Configuración más declarativa

### Ejemplo de Migración

**Antes:**
```json
{
  "nodes": [
    {"id": "task1", "type": "Task", "taskFunc": "Task1"}
  ],
  "edges": [
    {"from": "task1", "to": "task2"}
  ]
}
```

**Después:**
```json
{
  "start_node": "task1",
  "nodes": [
    {
      "id": "task1", 
      "type": "task", 
      "function": "Task1",
      "next": ["task2"]
    }
  ]
}
```

## Contribución

Para contribuir al proyecto:

1. Fork del repositorio
2. Crear branch de feature (`git checkout -b feature/nueva-caracteristica`)
3. Commit de cambios (`git commit -am 'Agregar nueva característica'`)
4. Push al branch (`git push origin feature/nueva-caracteristica`)
5. Crear Pull Request

## Licencia

MIT License - ver archivo LICENSE para detalles.
