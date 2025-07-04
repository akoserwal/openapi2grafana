# Grafana Dashboard Generator from OpenAPI

A powerful tool to automatically generate comprehensive Grafana dashboards from OpenAPI specifications with enhanced metrics, versioning, and monitoring capabilities.

## Features

- 🚀 **Enhanced Metrics**: P50, P90, P95, P99 percentiles, throughput, error rates
- 🔄 **Dashboard Versioning**: Track changes and update existing dashboards
- 📊 **Multiple Panel Types**: Time series, stats, and gauge panels
- 🎛️ **Advanced Templating**: Dynamic service, environment, and datasource variables
- 🔍 **gRPC Support**: Automatic detection and monitoring of gRPC services
- 🐳 **Docker Integration**: Complete monitoring stack with Prometheus and Grafana
- 🎨 **Modern UI**: Beautiful, responsive panels with proper thresholds
- 📈 **Alerting**: Built-in AlertManager integration
- 🔧 **Automation**: Watch mode for automatic regeneration

## Quick Start

### Prerequisites

- Go 1.18+ 
- Docker & Docker Compose
- jq (for JSON validation)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd gen-grafana-from-api

# Setup development environment
make dev-setup

# Check prerequisites
make check
```

### 🚀 Quick Demo with Sample API

We provide a complete sample API service that demonstrates all features:

```bash
# Option 1: One-command demo (recommended)
make demo

# Option 2: Manual setup
make generate-sample && make start-fresh

# Wait for services to start (about 30 seconds)
# Then visit:
# - Grafana: http://localhost:3000 (admin/admin)
# - Sample API: http://localhost:8080/api/inventory/v1/livez
# - Prometheus: http://localhost:9090
```

The sample API includes:
- ✅ **Full OpenAPI spec implementation** with 8+ endpoints
- ✅ **Real Prometheus metrics** (request rates, latency histograms, error rates)
- ✅ **Realistic behavior** with artificial latency and error simulation
- ✅ **Background traffic generation** for demonstration
- ✅ **Health checks** and proper service discovery

### Basic Usage

```bash
# Quick demo with sample API
make generate-sample && make start-fresh

# Generate dashboard from your OpenAPI spec  
make generate

# Start monitoring stack
make start

# Run full integration test with sample API
make test-full

# Clean up everything
make clean
```

## Advanced Usage

### Dashboard Generation Options

```bash
# Generate with custom options
go run main.go openapi.yaml dashboard.json --datasource prometheus --title "My API"

# Update existing dashboard
go run main.go openapi.yaml dashboard.json --update --uid my-dashboard

# Custom configuration
go run main.go openapi.yaml dashboard.json \
  --datasource prometheus \
  --title "Production API Dashboard" \
  --uid prod-api-dashboard
```

### Available Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Build the binary |
| `make generate` | Generate dashboard from OpenAPI spec |
| `make generate-sample` | Generate dashboard from sample API spec |
| `make update` | Update existing dashboard |
| `make build-sample-api` | Build sample API locally |
| `make run-sample-api` | Run sample API locally on :8080 |
| `make start` | Start monitoring stack |
| `make start-fresh` | Start monitoring stack with fresh build |
| `make stop` | Stop monitoring stack |
| `make test-full` | Run full integration test |
| `make validate` | Validate generated dashboard |
| `make watch` | Watch for changes and auto-regenerate |
| `make demo` | Run complete demo with sample API |
| `make clean` | Clean up all resources |

### Test Script Options

```bash
# Run basic test
./test-dashboard.sh

# Show logs after startup
./test-dashboard.sh --logs

# Clean up containers
./test-dashboard.sh --cleanup

# Show help
./test-dashboard.sh --help
```

## Dashboard Features

### Metrics Generated

For each API endpoint, the dashboard includes:

1. **Request Rate Panel**
   - Requests per second by status code
   - Time series visualization
   - Color-coded by HTTP status

2. **Latency Percentiles Panel**
   - P50, P90, P95, P99 response times
   - Time series with threshold alerts
   - Millisecond precision

3. **Error Rate Panel**
   - Percentage of 5xx errors
   - Stat panel with color thresholds
   - Real-time updates

4. **Throughput Panel**
   - Total requests per second
   - Stat panel with trends
   - Performance indicators

### Variables & Templating

- **Datasource**: Dynamic datasource selection
- **Environment**: Filter by environment (prod, stage, dev)
- **Service**: Filter by service name
- **Custom Variables**: Easily extensible

### gRPC Support

When gRPC extensions are detected in the OpenAPI spec:

```yaml
# In your OpenAPI spec
x-grpc:
  UserService:
    GetUser: {}
    CreateUser: {}
```

The tool automatically generates gRPC-specific panels with:
- gRPC status codes
- Method-specific latency
- Service-level metrics

## Configuration

### Prometheus Configuration

The generated dashboard expects these Prometheus metrics:

```yaml
# HTTP metrics
- http_requests_total{method, path, status_code, service}
- http_request_duration_seconds_bucket{method, path, service}

# gRPC metrics (if applicable)
- grpc_server_handled_total{grpc_service, grpc_method, grpc_code}
- grpc_server_handling_seconds_bucket{grpc_service, grpc_method}
```

### Customization

#### Adding Custom Metrics

Extend the `Panel` creation functions in `main.go`:

```go
func createCustomPanel(title, query string, panelID, height, yPos int) Panel {
    return Panel{
        ID:    panelID,
        Title: title,
        // ... panel configuration
    }
}
```

#### Custom Thresholds

Modify threshold values in the panel creation functions:

```go
Thresholds: ThresholdOptions{
    Mode: "absolute",
    Steps: []ThresholdStep{
        {Color: "green", Value: nil},
        {Color: "yellow", Value: floatPtr(0.5)},
        {Color: "red", Value: floatPtr(1.0)},
    },
},
```

## Monitoring Stack

### Services Included

- **Grafana**: Dashboard visualization (port 3000)
- **Prometheus**: Metrics collection (port 9090)
- **AlertManager**: Alert routing (port 9093)
- **Node Exporter**: System metrics (port 9100)

### Service URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | admin/admin |
| Prometheus | http://localhost:9090 | - |
| AlertManager | http://localhost:9093 | - |

### Data Persistence

- Grafana data: `grafana_data` volume
- Prometheus data: `prometheus_data` volume
- Dashboards: Auto-provisioned from `output_dashboard.json`

## Dashboard Structure

### Panel Layout

```
┌─────────────────────────────────────────────────────┐
│                Request Rate                         │
│                                                     │
├─────────────────────────────────────────────────────┤
│                Latency Percentiles                  │
│                                                     │
├─────────────────────────────────────────────────────┤
│ Error Rate          │         Throughput            │
│                     │                               │
└─────────────────────────────────────────────────────┘
```

### Versioning

The dashboard includes metadata for version tracking:

```json
{
  "meta": {
    "version": 2,
    "generated": "2024-01-01T12:00:00Z",
    "spec_hash": "abc123...",
    "last_updated": "2024-01-01T12:00:00Z"
  }
}
```

## Troubleshooting

### Common Issues

1. **Dashboard not importing**
   ```bash
   # Check Grafana logs
   docker-compose logs grafana
   
   # Validate dashboard JSON
   make validate
   ```

2. **No metrics appearing**
   ```bash
   # Check Prometheus targets
   curl http://localhost:9090/api/v1/targets
   
   # Verify metrics endpoint
   curl http://your-api:8080/metrics
   ```

3. **Permission issues**
   ```bash
   # Fix script permissions
   chmod +x test-dashboard.sh
   ```

### Debug Mode

Enable debug logging:

```bash
# Set environment variable
export DEBUG=true

# Run with verbose output
./test-dashboard.sh --logs
```

## Development

### Code Structure

```
main.go              # Main application logic
types.go            # Grafana dashboard types
panels.go           # Panel creation functions
docker-compose.yaml # Monitoring stack
prometheus.yml      # Prometheus configuration
test-dashboard.sh   # Integration test script
Makefile           # Build automation
```

### Adding New Panel Types

1. Define the panel structure in the types
2. Create a panel creation function
3. Add the panel to the dashboard generation logic
4. Update tests

### Testing

```bash
# Run unit tests
go test ./...

# Run integration tests
make test-full

# Manual testing
make generate && make start
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Run `make test-full` to verify
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Changelog

### v2.0.0
- ✨ Enhanced metrics with P90, P99 percentiles
- 🔄 Dashboard versioning and update mechanism
- 🎨 Modern panel designs with proper thresholds
- 🐳 Improved Docker Compose setup
- 📊 Better templating and variables
- 🔍 Enhanced gRPC support
- 🎯 Comprehensive test suite

### v1.0.0
- 🚀 Initial release with basic dashboard generation
- 📈 HTTP metrics support
- 🔧 Basic Docker integration 