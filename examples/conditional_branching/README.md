# Conditional Branching Workflow Example

This example demonstrates how to create workflows with conditional branching logic based on customer types and order values. The workflow classifies orders and routes them through different processing paths.

## 🎯 What This Example Shows

- **Conditional Logic**: Route workflow execution based on data conditions
- **Multiple Branches**: Handle different customer types (VIP, Regular, New, Invalid)
- **Data Classification**: Analyze and categorize input data
- **Three Implementation Modes**: Config-based, Programmatic, and Auto-discovery

## 🏗️ Workflow Architecture

```
┌─────────────────┐
│  Classify Order │
└─────────┬───────┘
          │
    ┌─────▼──────┐
    │ VIP Check? │
    └──────┬─────┘
           │
    ┌──────▼──────────────────┐
    │ Yes: Process VIP Order  │
    └─────────────────────────┘
           │
    ┌──────▼──────────────────┐
    │ No: Regular Check?      │
    └──────┬──────────────────┘
           │
    ┌──────▼─────────────────────┐
    │ Yes: Process Regular Order │
    └────────────────────────────┘
           │
    ┌──────▼──────────────────┐
    │ No: New Customer Check? │
    └──────┬──────────────────┘
           │
    ┌──────▼─────────────────────┐
    │ Yes: Process New Customer  │
    └────────────────────────────┘
           │
    ┌──────▼──────────────────────┐
    │ No: Handle Invalid Order    │
    └─────────────────────────────┘
           │
    ┌──────▼──────────┐
    │ Finalize Order  │
    └─────────────────┘
```

## 🚀 Usage Examples

### 1. Config-based Mode (Default)
```bash
# Process regular customer order
cd examples/conditional_branching
go run main.go -mode=config -customer=regular -order=150

# Process VIP customer order
go run main.go -mode=config -customer=vip -order=500

# Process new customer order
go run main.go -mode=config -customer=new -order=75

# Handle invalid customer type
go run main.go -mode=config -customer=invalid -order=100
```

### 2. Programmatic Mode
```bash
# Build workflow entirely in code
go run main.go -mode=programmatic -customer=vip -order=300
```

### 3. Auto-discovery Mode
```bash
# Auto-discover tasks and generate workflow
go run main.go -mode=autodiscovery -customer=regular -order=200
```

## 📋 Workflow Tasks

| Task | Description | Input | Output |
|------|-------------|--------|---------|
| `classifyOrder` | Analyzes customer type and order value | Order data | Classification result |
| `processVIPOrder` | Handles VIP customers with premium benefits | Classified data | VIP processing result |
| `processRegularOrder` | Processes regular customer orders | Classified data | Regular processing result |
| `processNewCustomer` | Handles new customers with welcome benefits | Classified data | New customer result |
| `handleInvalidOrder` | Manages invalid orders and escalates to support | Classified data | Error handling result |
| `finalizeOrder` | Prepares order for shipment and tracking | Processed data | Final order status |

## 🎨 Customer Classification Logic

### VIP Customers (`customer_type: "vip"`)
- **Priority**: 1 (Highest)
- **Benefits**: 15% discount, expedited shipping, priority support, 500 loyalty points
- **Processing Time**: ~100ms

### Regular Customers (`customer_type: "regular"`)
- **High Value** (≥$100): 5% discount, 100 loyalty points
- **Low Value** (<$100): 0% discount, 50 loyalty points  
- **Priority**: 2-3
- **Processing Time**: ~50ms

### New Customers (`customer_type: "new"`)
- **Priority**: 4
- **Benefits**: 10% welcome discount, 200 bonus points, welcome email
- **Processing Time**: ~75ms

### Invalid Orders
- **Priority**: 5
- **Actions**: Create support ticket, escalate to support team
- **Processing Time**: ~25ms

## 📁 File Structure

```
conditional_branching/
├── main.go                 # Main entry point with CLI
├── README.md              # This documentation
├── config/
│   └── workflow.json      # Workflow configuration
├── programmatic/
│   └── workflow.go        # Programmatic workflow builder
└── tasks/
    ├── annotations.go     # Task annotations for auto-discovery
    └── tasks_clean.go     # All workflow task implementations
```

## 🔧 Configuration Options

### Command Line Flags
- `-mode`: Execution mode (`config`, `programmatic`, `autodiscovery`)
- `-customer`: Customer type (`vip`, `regular`, `new`, `invalid`)
- `-order`: Order value (float64)

### Example Configurations

```bash
# High-value regular customer
go run main.go -customer=regular -order=250

# Low-value regular customer  
go run main.go -customer=regular -order=50

# VIP customer with large order
go run main.go -customer=vip -order=1000

# New customer with medium order
go run main.go -customer=new -order=125
```

## 📊 Expected Output

```
🚀 Starting Conditional Branching Example
📋 Mode: config
👤 Customer Type: vip
💰 Order Value: $500.00

⚙️ Running in CONFIG mode...
🔍 Classifying order...
   ✅ Order classified as: vip (priority: 1)
👑 Processing VIP order...
   ✅ VIP order processed with premium benefits
🎯 Finalizing order...
   ✅ Order finalized and ready for shipment (Type: vip)
✅ Workflow completed successfully!

🎉 Conditional Branching completed!
📈 Execution Details:
  - Node ID: finalizeOrder
  - Success: true
  - Timestamp: 1642789234

📋 Final Result:
{
  "original_data": {...},
  "finalized_at": 1642789234,
  "processing_type": "vip",
  "final_status": "finalized",
  "ready_for_shipment": true,
  "tracking_number": "TRK_1642789234",
  "estimated_delivery": "2025-07-24"
}
```

## 🧪 Testing

The example includes comprehensive testing scenarios for all customer types and order values. Each branch of the conditional logic is validated to ensure proper routing and processing.

## 🔗 Related Examples

- **Basic Workflow**: Simple linear workflow execution
- **Parallel Processing**: Concurrent task execution
- **Data Processing Pipeline**: Complex data transformations
- **Advanced Hooks & Monitoring**: Workflow observability
