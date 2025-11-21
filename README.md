# API Gateway

A production-ready API Gateway microservice built with Go, Gin, and the `go-service-kit` common library. This gateway handles request routing, authentication, trace propagation, and intelligent alerting.

## Features

- **Request Routing**: Proxies requests to downstream microservices (e.g., `/hrms/*` → `hrms-core`)
- **Trace Propagation**: Generates and propagates `X-Trace-ID` and `X-Request-ID` headers
- **Intelligent Alerting**: Only alerts on errors/slow requests if downstream service hasn't already handled them
- **Telemetry Integration**: New Relic and Slack webhook support
- **Retry Logic**: Automatic retry with exponential backoff for downstream requests
- **Structured Logging**: Context-aware logging with trace and request IDs

## Configuration

Configuration is loaded from `config.yaml` with environment variable overrides. Key settings:

### Environment Variables

- `GATEWAY_PORT`: HTTP server port (default: 8080)
- `GATEWAY_SLOW_MS`: Latency threshold in milliseconds for slow request detection (default: 1000)
- `TIMEOUT_MS`: Request timeout in milliseconds (default: 5000)
- `HRMS_CORE_BASE_URL`: Base URL for hrms-core service (default: http://localhost:8081)
- `NEWRELIC_LICENSE_KEY`: New Relic license key
- `SLACK_WEBHOOK_URL`: Slack webhook URL for alerts
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `LOG_FORMAT`: Log format (json, text)

### Config File

See `config.yaml` for the full configuration structure.

## Running Locally

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Update config.yaml** with your telemetry credentials (or use environment variables)

3. **Run the service:**
   ```bash
   go run cmd/api-gateway/main.go
   ```

## Building

```bash
go build -o api-gateway ./cmd/api-gateway
```

## Docker

```bash
docker build -t api-gateway .
docker run -p 8080:8080 \
  -e NEWRELIC_LICENSE_KEY=your-key \
  -e SLACK_WEBHOOK_URL=your-webhook \
  api-gateway
```

## API Endpoints

### Health Check
```
GET /health
```

### HRMS Routes
All routes under `/hrms/*` and `/employees/*` are proxied to the `hrms-core` service.

## Example Requests

### Get Employee (via Gateway)
```bash
curl -X GET http://localhost:8080/hrms/v1/employees/1 \
  -H "X-Trace-ID: $(uuidgen)"
```

### Test Slow Request
```bash
curl -X GET http://localhost:8080/hrms/v1/employees/slow
```

### Test Error Handling
```bash
curl -X GET http://localhost:8080/hrms/v1/employees/error
```

## Trace Propagation

The gateway:
1. Generates a `X-Trace-ID` (UUID v4) for each client request
2. Generates a `X-Request-ID` (gateway request ID)
3. Forwards both headers to downstream services
4. Includes both IDs in all logs and telemetry events

## Alerting Logic

The gateway triggers alerts (New Relic + Slack) when:
- **Slow Request**: Latency > `GATEWAY_SLOW_MS` AND downstream service didn't handle it
- **Gateway Error**: HTTP 5xx status AND downstream service didn't handle it

The gateway **does NOT** alert if:
- Downstream response includes `X-Service-Handled: true` header
- Downstream error response has `handled_by_service: true` in JSON payload

## Response Format

Error responses follow the `go-service-kit` standard format:

```json
{
  "code": "INTERNAL_ERROR",
  "message": "user-friendly message",
  "details": {
    "operation": "getEmployee"
  },
  "handled_by_service": true
}
```

## Testing

```bash
# Run tests
go test ./...

# Integration test
go test ./test/...
```

## Dependencies

- `github.com/yourorg/go-service-kit`: Shared service library
- `github.com/gin-gonic/gin`: HTTP framework
- `github.com/newrelic/go-agent/v3`: New Relic agent

