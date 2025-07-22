# Data Processing Pipeline Example

Este ejemplo demuestra un pipeline real de procesamiento de datos:

## 🎯 Lo que demuestra:
- HTTP + Foreach combinados
- Validación de datos en lote
- Transformaciones complejas
- Manejo robusto de errores
- Procesamiento paralelo

## 📋 Flujo del Workflow:
```
Fetch Users API → Validate Each User → Transform Data → Save to DB → Generate Report
```

## 🚀 Cómo ejecutar:

### Por código:
```bash
cd examples/data_processing_pipeline
go run main.go --mode=programmatic --source=api
```

### Por archivo de configuración:
```bash
go run main.go --mode=config --source=mock
```

### Con auto-discovery:
```bash
go run main.go --mode=autodiscovery --source=api
```

## 🌐 Fuentes de datos:
- **API**: JSONPlaceholder (usuarios reales)
- **Mock**: Datos de prueba locales

## 📊 Datos de entrada:
```json
{
  "api_url": "https://jsonplaceholder.typicode.com/users",
  "batch_size": 5,
  "processing_rules": {
    "validate_email": true,
    "normalize_names": true,
    "extract_domains": true
  }
}
```

## 🔄 Resultado esperado:
```json
{
  "total_users": 10,
  "valid_users": 8,
  "processed_users": 8,
  "failed_validations": 2,
  "domains_found": ["example.com", "gmail.com", "yahoo.com"],
  "processing_time": "2.34s",
  "report_generated": true
}
```
