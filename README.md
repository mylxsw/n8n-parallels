# N8n Parallels Server

N8n Parallels Server is a web service designed to execute webhook requests in parallel for N8N workflows. It receives a webhook URL and multiple payloads, executes HTTP requests to the webhook in parallel, and returns the aggregated results while maintaining the original order.

## Features

- **Parallel Execution**: Execute multiple webhook requests concurrently
- **Timeout Support**: Configurable timeout for each request with automatic handling
- **Order Preservation**: Results are returned in the same order as the input payloads
- **Comprehensive Logging**: Structured logging with configurable levels
- **Health Checks**: Built-in health check endpoint
- **Graceful Shutdown**: Proper cleanup on termination signals
- **Docker Support**: Ready-to-use Docker image

## API Documentation

### Execute Parallel Webhooks

**Endpoint:** `POST /v1/parallels/execute`

**Request Body:**
```json
{
    "webhook_url": "https://your-webhook-endpoint.com/webhook",
    "auth_header": "Bearer your-token-here",
    "payloads": [
        {
            "arg1": "val1",
            "arg2": "val2"
        },
        {
            "arg1": "val3",
            "arg2": "val4"
        }
    ],
    "timeout": 60
}
```

**Request Parameters:**
- `webhook_url` (string, required): The webhook URL to send requests to
- `auth_header` (string, optional): Authorization header value (e.g., "Bearer token")
- `payloads` (array, required): Array of objects, each will be sent as a separate HTTP request
- `timeout` (int, optional): Timeout in seconds for each request (default: 60, max: 3600)

**Response:**
```json
{
    "results": [
        {
            "index": 0,
            "success": true,
            "response": {"result": "success"},
            "duration_ms": 150
        },
        {
            "index": 1,
            "success": false,
            "error": "timeout",
            "duration_ms": 60000
        }
    ],
    "summary": {
        "total_requests": 2,
        "successful_requests": 1,
        "failed_requests": 1,
        "timeout_requests": 1,
        "total_duration_ms": 60200
    }
}
```

**Response Fields:**
- `results`: Array of individual webhook execution results
  - `index`: Position in the original payloads array
  - `success`: Whether the request succeeded (2xx status code)
  - `response`: Raw response body (only present on success)
  - `error`: Error message (only present on failure)
  - `duration_ms`: Request duration in milliseconds
- `summary`: Execution summary statistics

### Health Check

**Endpoint:** `GET /health`

**Response:**
```json
{
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "service": "n8n-parallels",
    "version": "1.0.0"
}
```

## Quick Start

### Using Go

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd n8n-parallels
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run the application:
   ```bash
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:8080` by default.

### Using Docker

1. Build the Docker image:
   ```bash
   docker build -t n8n-parallels .
   ```

2. Run the container:
   ```bash
   docker run -p 8080:8080 n8n-parallels
   ```

### Using Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3.8'
services:
  n8n-parallels:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - LOG_LEVEL=info
      - LOG_FORMAT=json
    restart: unless-stopped
```

Then run:
```bash
docker-compose up -d
```

## Configuration

The application can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `HOST` | `0.0.0.0` | Server host |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `LOG_FORMAT` | `text` | Log format (text, json) |
| `READ_TIMEOUT` | `30` | HTTP read timeout in seconds |
| `WRITE_TIMEOUT` | `30` | HTTP write timeout in seconds |
| `SHUTDOWN_TIMEOUT` | `30` | Graceful shutdown timeout in seconds |

## Usage Examples

### Basic Usage

```bash
curl -X POST http://localhost:8080/v1/parallels/execute \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "https://httpbin.org/post",
    "payloads": [
      {"id": 1, "name": "Alice"},
      {"id": 2, "name": "Bob"},
      {"id": 3, "name": "Charlie"}
    ],
    "timeout": 30
  }'
```

### With Authentication

```bash
curl -X POST http://localhost:8080/v1/parallels/execute \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "https://your-api.com/webhook",
    "auth_header": "Bearer your-jwt-token",
    "payloads": [
      {"action": "create", "data": {"name": "Item 1"}},
      {"action": "create", "data": {"name": "Item 2"}}
    ],
    "timeout": 60
  }'
```

### Health Check

```bash
curl http://localhost:8080/health
```

## Integration with N8N

In your N8N workflow:

1. Add an HTTP Request node
2. Set the URL to your N8n Parallels Server: `http://your-server:8080/v1/parallels/execute`
3. Set method to POST
4. Configure the request body with your webhook URL and payloads
5. The node will wait for all parallel requests to complete and return the aggregated results

## Development

### Project Structure

```
n8n-parallels/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP request handlers
│   ├── logger/          # Logging configuration
│   ├── models/          # Data models
│   └── service/         # Business logic
├── Dockerfile           # Docker image definition
├── docker-compose.yml   # Docker Compose configuration
├── go.mod              # Go module definition
└── README.md           # This file
```

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o n8n-parallels ./cmd/server
```

## Performance Considerations

- The service uses goroutines for parallel execution, providing excellent concurrency
- Memory usage scales with the number of concurrent requests
- Default timeouts are conservative; adjust based on your webhook response times
- Consider resource limits in containerized environments

## Error Handling

The service handles various error scenarios:

- **Network errors**: Connection failures, DNS resolution errors
- **Timeouts**: Requests exceeding the specified timeout
- **HTTP errors**: Non-2xx response codes from webhooks
- **Invalid payloads**: Malformed JSON or validation errors

All errors are properly logged and returned in a structured format.

## License

[Add your license information here]

## Contributing

[Add contributing guidelines here]

## Support

[Add support information here]