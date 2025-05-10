# PadelAlert

A service to notify you when a Playtomic padel match or class is available for a particular club, ranking range, or other criteria.

## Features

- Monitor available padel matches and classes on Playtomic
- Filter matches by club, ranking, date range, and more
- Send notifications via Email (with future support for Telegram and SMS)
- RESTful API for managing notification rules
- Health check endpoint for monitoring
- Simple and efficient architecture
- Performance-optimized design

## API Endpoints

### Authentication

The API uses API key authentication for protected endpoints. To access protected endpoints, you need to include an API key in your request using one of the following methods:

1. **HTTP Header**: Include the API key in the `X-API-Key` header:
   ```
   X-API-Key: your_api_key_here
   ```

2. **Query Parameter**: Include the API key as a query parameter:
   ```
   ?api_key=your_api_key_here
   ```

API keys are configured in the `.env` file using the `API_KEYS` environment variable, which accepts a comma-separated list of valid API keys.

### Status Endpoints

- `GET /api/v1/health`: Health check endpoint (public)
- `GET /metrics`: Prometheus metrics endpoint (protected)

### Rule Management Endpoints

- `GET /api/v1/rules`: List all notification rules (protected)
- `GET /api/v1/rules/<rule_id>`: Get a specific rule (protected)
- `POST /api/v1/rules`: Create a new rule (protected)
- `PUT /api/v1/rules/<rule_id>`: Update a rule (protected)
- `DELETE /api/v1/rules/<rule_id>`: Delete a rule (protected)

### Admin Endpoints

- `GET /admin/notifications`: List all notifications (protected)
- `POST /admin/clear-notifications`: Clear all notifications (protected)

## Creating a Rule

To create a notification rule, send a POST request to `/api/v1/rules`:

```json
{
  "rule_type": "match",
  "name": "My Match Rule",
  "club_ids": ["12345"],
  "user_id": "your-user-id",
  "email": "your.email@example.com",
  "min_ranking": 3.0,
  "max_ranking": 4.5,
  "start_date": "2023-01-01",
  "end_date": "2023-01-31"
}
```

For a class rule:

```json
{
  "rule_type": "class",
  "name": "My Class Rule",
  "club_ids": ["12345"],
  "user_id": "your-user-id",
  "email": "your.email@example.com",
  "title_contains": "beginner"
}
```

Note: The `user_id` and `email` fields are required to identify who should receive notifications.

## Configuration

PadelAlert is configured through environment variables, typically stored in a `.env` file:

```
# HTTP server settings
PORT=8080
API_KEYS=key1,key2
LOG_LEVEL=info
API_RATE_LIMIT=10

# Redis settings
REDIS_URL=redis://localhost:6379

# Scheduler configuration
CHECK_INTERVAL=300

# Email settings
SMTP_SERVER=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=user@example.com
SMTP_PASSWORD=yourpassword
SMTP_SENDER=alerts@example.com
```

See `.env.example` for a complete list of available configuration options.

## Deployment

### Using Docker

```bash
# Build the image
docker build -t padel-alert .

# Run with environment variables
docker run -p 8080:8080 --env-file .env padel-alert:latest
```

## Development

### Prerequisites

- Go 1.24 or higher
- Redis
- Docker (optional)

### Getting Started

1. Clone the repository
2. Copy `.env.example` to `.env` and configure the variables
3. Start Redis: `docker run -d -p 6379:6379 redis:alpine`
4. Start the server: `go run cmd/padel-alert/main.go server`
5. Visit http://localhost:8080/api/v1/health to check if the server is running

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run performance benchmarks
go test -bench=. -benchmem ./...
```

### Performance Testing

The repository includes tools for performance testing and optimization:

```bash
# Run performance benchmarks
go run tools/performance/benchmark.go

# Generate CPU profile
go run tools/performance/benchmark.go -cpuprofile=cpu.prof

# Generate memory profile
go run tools/performance/benchmark.go -memprofile=mem.prof

# Analyze profiles
go tool pprof -http=:8080 cpu.prof
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
