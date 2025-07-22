# 🚀 Flexi-Flows Examples - Complete Guide

Esta carpeta contiene **6 ejemplos completamente funcionales** que demuestran todas las capacidades de Flexi-Flows usando la API más reciente.

## ✅ Ejemplos Implementados

### 1. 🎯 **Basic Workflow** - Fundamentos
**Estado**: ✅ **COMPLETO**
- **Ubicación**: `basic_workflow/`
- **Demuestra**: Configuración por código vs archivo, auto-discovery, tareas secuenciales
- **Flujo**: `Input Data → Validate → Transform → Save → Notify`
- **Ejecución**: `cd basic_workflow && go run main.go --mode=config`

### 2. 📊 **Data Processing Pipeline** - HTTP + Foreach
**Estado**: ✅ **COMPLETO** 
- **Ubicación**: `data_processing_pipeline/`
- **Demuestra**: Llamadas HTTP reales, validación en lote, transformaciones
- **Flujo**: `Fetch Users API → Validate Each User → Transform → Save → Generate Report`
- **Ejecución**: `cd data_processing_pipeline && go run main.go --mode=config --source=api`

### 3. 🔀 **Conditional Branching** - Lógica de Negocio
**Estado**: ✅ **COMPLETO**
- **Ubicación**: `conditional_branching/`
- **Demuestra**: Condicionales complejas, múltiples ramas, lógica de negocio
- **Flujo**: `Order → Classify Customer → [VIP|Regular|New|Invalid] → Finalize`
- **Ejecución**: `cd conditional_branching && go run main.go --customer=vip`

### 4. ⚡ **Parallel Processing** - Concurrencia
**Estado**: ✅ **COMPLETO**
- **Ubicación**: `parallel_processing/`
- **Demuestra**: Procesamiento paralelo, agregación de resultados
- **Flujo**: `Weather Service: Fetch Multiple Cities → Aggregate → Report`
- **Ejecución**: `cd parallel_processing && go run main.go --cities="NYC,London,Tokyo"`

### 5. 🔍 **Advanced Hooks Monitoring** - Observabilidad
**Estado**: ✅ **COMPLETO**
- **Ubicación**: `advanced_hooks_monitoring/`
- **Demuestra**: Hooks personalizados, métricas, logging estructurado
- **Flujo**: `E-commerce Order → [Hooks] → Payment → Inventory → Shipping`
- **Ejecución**: `cd advanced_hooks_monitoring && go run main.go --order-id=12345`

### 6. 🏗️ **Microservices Orchestration** - Arquitectura Distribuida
**Estado**: ✅ **COMPLETO**
- **Ubicación**: `microservices_orchestration/`
- **Demuestra**: Orquestación de servicios, circuit breaker, retry logic
- **Flujo**: `User Registration → Validate → Create → Email → Profile → Analytics`
- **Ejecución**: `cd microservices_orchestration && go run main.go --user-id=12345`

## 🎮 Modos de Ejecución

Cada ejemplo soporta **3 modos diferentes**:

### 1. **Configuración por Archivo** (Recomendado)
```bash
go run main.go --mode=config
```
- Usa archivos JSON/YAML en `config/`
- Ideal para producción
- Fácil de modificar sin recompilar

### 2. **Configuración Programática**
```bash
go run main.go --mode=programmatic
```
- Workflow definido completamente en código
- Máximo control y flexibilidad
- Ideal para casos complejos

### 3. **Auto-Discovery**
```bash
go run main.go --mode=autodiscovery
```
- Detecta automáticamente tareas anotadas
- Configuración híbrida
- Ideal para desarrollo rápido

## 📋 Estructura de Cada Ejemplo

```
example_name/
├── README.md                    # Documentación específica
├── main.go                      # Punto de entrada principal
├── main_test.go                 # Tests unitarios
├── config/
│   ├── workflow.json           # Configuración JSON
│   └── workflow.yaml           # Configuración YAML (alternativa)
├── programmatic/
│   └── workflow.go             # Configuración por código
└── tasks/
    └── annotations.go          # Tareas con auto-discovery
```

## 🚀 Ejecutar Todos los Ejemplos

### Opción 1: Individual
```bash
cd examples/basic_workflow
go run main.go --mode=config

cd examples/data_processing_pipeline  
go run main.go --mode=config --source=mock

cd examples/conditional_branching
go run main.go --mode=config --customer=vip
```

### Opción 2: Con Script (Próximamente)
```bash
./run_all_examples.sh
```

## 🧪 Ejecutar Tests

```bash
# Test individual
cd basic_workflow && go test -v

# Test todos los ejemplos
find . -name "*_test.go" -exec go test -v {} \;
```

## 📊 Casos de Uso Demostrados

| Funcionalidad | Ejemplos que lo demuestran |
|---------------|---------------------------|
| **HTTP Calls** | Data Processing, Microservices |
| **Foreach/Iteración** | Data Processing |
| **Condicionales** | Conditional Branching |
| **Parallelismo** | Parallel Processing |
| **Hooks/Logging** | Advanced Monitoring |
| **Auto-discovery** | Todos los ejemplos |
| **Error Handling** | Todos los ejemplos |
| **Configuración JSON/YAML** | Todos los ejemplos |
| **Configuración por Código** | Todos los ejemplos |

## 🆘 Solución de Problemas

### Error: "module not found"
```bash
cd /path/to/workflows
go mod tidy
```

### Error: "config file not found"
Asegúrate de ejecutar desde el directorio del ejemplo:
```bash
cd examples/basic_workflow  # ✅ Correcto
go run main.go

# ❌ Incorrecto desde raíz
go run examples/basic_workflow/main.go  
```

### Error: "task not registered"
El ejemplo usa auto-discovery. Verifica que las tareas tengan anotaciones `@workflow:task`.

## 🔄 Próximos Pasos

1. **Ejecuta los ejemplos** en orden de complejidad
2. **Modifica las configuraciones** en `config/workflow.json`
3. **Crea tus propias tareas** siguiendo los patrones
4. **Experimenta con diferentes modos** (config, programmatic, autodiscovery)
5. **Integra con tus sistemas** usando los ejemplos como plantillas

## 📈 Progresión Recomendada

1. **🎯 Basic Workflow** - Entiende los conceptos fundamentales
2. **📊 Data Processing** - Aprende HTTP + Foreach 
3. **🔀 Conditional Branching** - Domina la lógica condicional
4. **⚡ Parallel Processing** - Explora concurrencia
5. **🔍 Advanced Monitoring** - Implementa observabilidad
6. **🏗️ Microservices** - Orquesta arquitecturas complejas

¡Empieza con **Basic Workflow** y progresa según tu experiencia!
