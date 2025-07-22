# Conditional Branching Example

Este ejemplo demuestra lógica de negocio compleja con ramificaciones condicionales:

## 🎯 Lo que demuestra:
- Condicionales complejas
- Múltiples ramas de ejecución
- Branch + Merge patterns
- Lógica de negocio real

## 📋 Flujo del Workflow:
```
Order Processing:
├─ VIP Customer? → Priority Queue → Express Shipping
├─ Regular Customer? → Standard Queue → Normal Shipping  
├─ New Customer? → Verification → Welcome Process
└─ Invalid Order? → Error Handling → Notification
```

## 🚀 Cómo ejecutar:

```bash
cd examples/conditional_branching
go run main.go --mode=config --customer=vip
go run main.go --mode=programmatic --customer=regular
go run main.go --mode=autodiscovery --customer=new
```

## 📊 Tipos de clientes:
- **VIP**: Procesamiento prioritario
- **Regular**: Procesamiento estándar
- **New**: Verificación adicional
- **Invalid**: Manejo de errores
