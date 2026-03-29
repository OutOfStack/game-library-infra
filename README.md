# game-library-infra

Infrastructure services for [game-library](https://github.com/OutOfStack/game-library) application. Includes distributed tracing, log management, and metrics collection.

## Services

| Service | Description |
|---------|-------------|
| **Zipkin** | Distributed tracing system |
| **Jaeger** | Distributed tracing system with native OTLP support |
| **Graylog** | Log management platform (with MongoDB and OpenSearch) |
| **Prometheus** | Metrics collection and monitoring |
| **Grafana** | Metrics visualization and dashboards |

## Quick Start with Docker

```sh
# Start all services
make all

# Or start individual services
make zipkin
make jaeger
make prometheus
make graylog
make grafana

# Stop all services
make stop

# Stop and remove volumes
make clean
```

## Access URLs and Credentials

| Service | URL | Credentials |
|---------|-----|-------------|
| Zipkin | http://localhost:9411/zipkin | No authentication required |
| Jaeger | http://localhost:16686 | No authentication required |
| Prometheus | http://localhost:9090 | No authentication required |
| Graylog | http://localhost:9000 | `admin:admin` |
| OpenSearch | http://localhost:9200 | No authentication (security disabled) |
| Grafana | http://localhost:3000 | `admin:admin` |

## Graylog Setup

After starting Graylog, you need to create a GELF input to receive logs:

1. Open Graylog at http://localhost:9000 and log in with `admin`/`admin`
2. Go to **System** → **Inputs**
3. Select **GELF UDP** from the dropdown and click **Launch new input**
4. Configure the input:
   - **Global**: Check this box
   - **Title**: `GELF UDP`
   - **Bind address**: `0.0.0.0`
   - **Port**: `12201`
5. Click **Launch** and complete the setup wizard

The input will now accept GELF messages on UDP port 12201.
## Example Application

The `example/` folder contains a sample Go application demonstrating integration with all infrastructure services.

### Features

- **Distributed tracing** - OpenTelemetry with Zipkin exporter and OTLP exporter (Jaeger)
- **Graylog logging** - zap logger with GELF output
- **Prometheus metrics** - request counters, error counters, response duration histograms
- **Connected traces** - HTTP client and server spans are linked via trace propagation

### Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/data` | Returns in 10-100ms (random delay) |
| `GET /api/slow` | Returns in 200-1000ms (random delay), 10% error rate |
| `GET /metrics` | Prometheus metrics |
| `GET /health` | Health check |

### Running the Example

```sh
# Start infrastructure services first
make all

# Run the example app
make run-example
```

### Project Structure

```
game-library-infra/
├── .github/workflows/
│   └── main.yml          # CI/CD pipeline
├── .k8s/
│   ├── namespace.yaml    # Kubernetes namespace
│   ├── cert.yaml         # Let's Encrypt ClusterIssuers (HTTP-01 & DNS-01)
│   ├── secrets.yaml      # Secrets for MongoDB, Graylog, Cloudflare
│   ├── graylog.yaml      # Graylog deployment
│   ├── mongo.yaml        # MongoDB deployment
│   ├── opensearch.yaml   # OpenSearch deployment
│   ├── jaeger.yaml       # Jaeger deployment
│   ├── prometheus.yaml   # Prometheus deployment
│   └── zipkin.yaml       # Zipkin deployment (deprecated)
├── example/
│   ├── main.go           # Application entry point
│   ├── logger.go         # Zap + Graylog logger setup
│   ├── tracer.go         # OpenTelemetry tracer setup (Zipkin + OTLP exporters)
│   ├── metrics.go        # Prometheus metrics middleware
│   ├── handlers.go       # HTTP handlers
│   ├── client.go         # HTTP client with tracing + load generator
│   ├── go.mod
│   └── go.sum
├── docker-compose.yml    # Local development setup
├── Makefile              # Build and run commands
```
