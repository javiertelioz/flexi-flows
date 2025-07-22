# Parallel Processing - Weather Data Example

Este ejemplo demuestra el procesamiento paralelo de datos del clima usando Flexi-Flows.

## Funcionalidad

- **Obtención de datos del clima**: Simula llamadas paralelas a APIs para obtener datos meteorológicos de múltiples ciudades
- **Agregación de datos**: Procesa y agrega los datos obtenidos para generar estadísticas resumidas
- **Múltiples modos de ejecución**: Configuración por archivos, programático y auto-discovery

## Estructura

```
parallel_processing/
├── main.go                 # Punto de entrada principal
├── config/
│   └── workflow.json      # Configuración declarativa del workflow
├── tasks/
│   └── annotations.go     # Funciones de tareas con anotaciones
└── programmatic/
    └── workflow.go        # Implementación programática
```

## Uso

### Modo Configuración (recomendado)
```bash
go run main.go -mode=config -cities="Madrid,Barcelona,Valencia,Sevilla"
```

### Modo Programático
```bash
go run main.go -mode=programmatic -cities="New York,London,Tokyo,Paris"
```

## Salida Esperada

```
🚀 Starting Weather Processing Example
📋 Mode: config
🌍 Cities: Madrid,Barcelona,Valencia,Sevilla

⚙️ Running in CONFIG mode...
🌦️ Starting weather data fetch...
📍 Fetched weather for Madrid: 25.3°C, sunny
📍 Fetched weather for Barcelona: 28.7°C, cloudy
📍 Fetched weather for Valencia: 31.2°C, sunny
📍 Fetched weather for Sevilla: 35.1°C, sunny
✅ Successfully fetched weather data for 4 cities
📊 Starting weather data aggregation...
📊 Aggregated data for 4 cities:
   - Average Temperature: 30.1°C
   - Average Humidity: 65.2%
   - Most Common Condition: sunny
✅ Config workflow completed successfully!

🎉 Workflow completed successfully!
📈 Execution Details:
  - Node ID: aggregate_weather
  - Success: true
  - Timestamp: 1642781234

📋 Final Result:
{
  "aggregated_data": {
    "average_temperature": 30.1,
    "average_humidity": 65.2,
    "condition_summary": "sunny",
    "total_cities": 4
  },
  "weather_data": [...]
}
```

## Características Técnicas

- **Simulación de APIs**: Las funciones simulan llamadas a APIs externas con delays realistas
- **Procesamiento de datos**: Agregación matemática de temperatura, humedad y condiciones climáticas
- **Manejo de errores**: Validación robusta de datos de entrada y salida
- **Flexibilidad**: Configuración fácil del número y nombres de ciudades
