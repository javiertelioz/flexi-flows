# Microservices Orchestration Example

Demuestra orquestación de microservicios con manejo robusto de errores:

## 🎯 Características:
- HTTP calls a múltiples servicios
- Circuit breaker simulation
- Retry logic manual
- Timeout handling

## 📋 Flujo:
```
User Registration:
├─ Validate Data → Create User → Send Email → Create Profile → Update Analytics
```

## 🚀 Ejecutar:
```bash
cd examples/microservices_orchestration
go run main.go --mode=config --user-id=12345
```
