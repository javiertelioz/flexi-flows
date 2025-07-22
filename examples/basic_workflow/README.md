# Basic Workflow Example

Este ejemplo demuestra los conceptos fundamentales de Flexi-Flows:

## 🎯 Lo que demuestra:
- Configuración por código vs archivo
- Auto-discovery básico  
- Tareas secuenciales simples
- Validación y transformación de datos

## 📋 Flujo del Workflow:
```
Input Data → Validate → Transform → Save → Notify
```

## 🚀 Cómo ejecutar:

### Por código:
```bash
go run main.go --mode=programmatic
```

### Por archivo de configuración:
```bash
go run main.go --mode=config
```

### Con auto-discovery:
```bash
go run main.go --mode=autodiscovery
```

## 📊 Datos de ejemplo:
```json
{
  "user": {
    "name": "John Doe", 
    "email": "john@example.com",
    "age": 30
  }
}
```

## 🔄 Resultado esperado:
```json
{
  "validated": true,
  "transformed": {
    "fullName": "JOHN DOE",
    "email": "john@example.com", 
    "category": "adult"
  },
  "saved": true,
  "notified": true
}
```
