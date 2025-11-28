# Environment Configuration Guide

This document describes how to configure Zentube for different environments (development, staging, production).

## Environment Setup

Zentube supports environment-specific configuration through:
1. Environment variable `APP_ENV` 
2. Environment-specific config files (`config.<env>.yaml`)
3. Environment-specific .env files (`.env.<env>`)

### Supported Environments

- **development** (default) - Local development with debug logging
- **staging** - Pre-production testing environment
- **production** - Production deployment with optimized settings

## Configuration Priority

Configuration is loaded with the following priority:

1. **Environment Detection**: `APP_ENV` environment variable determines the active environment
2. **Config File Loading**: 
   - First tries `config.<env>.yaml` (e.g., `config.production.yaml`)
   - Falls back to `config.yaml` if environment-specific file doesn't exist
3. **Environment Variables**:
   - First tries `.env.<env>` (e.g., `.env.production`)
   - Falls back to `.env` if environment-specific file doesn't exist
4. **Secret Injection**: Sensitive values like API keys are injected from environment variables

## Environment-Specific Files

### Development

**File**: `configs/config.development.yaml`
```yaml
app:
  name: "zentube"
  port: 8080

youtube:
  api_key: "" # Injected from .env
  max_results: 10

database:
  path: "./zentube_dev.db"
```

**Environment Variables** (`.env` or `.env.development`):
```bash
YOUTUBE_API_KEY=your_dev_api_key_here
APP_ENV=development
```

**Features**:
- Debug mode enabled in Gin
- Text-based logging for better readability
- SQLite database in current directory
- Lower max results to save quota during development

### Staging

**File**: `configs/config.staging.yaml`
```yaml
app:
  name: "zentube"
  port: 8080

youtube:
  api_key: "" # Injected from .env.staging
  max_results: 15

database:
  path: "./zentube_staging.db"
```

**Environment Variables** (`.env.staging`):
```bash
YOUTUBE_API_KEY=your_staging_api_key_here
APP_ENV=staging
```

**Features**:
- Test mode in Gin
- JSON logging for structured logs
- Staging database
- Moderate max results

### Production

**File**: `configs/config.production.yaml`
```yaml
app:
  name: "zentube"
  port: 8080

youtube:
  api_key: "" # Injected from YOUTUBE_API_KEY env var
  max_results: 25

database:
  path: "/var/lib/zentube/zentube.db"
```

**Environment Variables** (System environment or `.env.production`):
```bash
YOUTUBE_API_KEY=your_production_api_key_here
APP_ENV=production
```

**Features**:
- Release mode in Gin (optimized, no debug logging)
- JSON logging for log aggregation
- Database in system directory
- Higher max results for better user experience
- In production, .env file is optional (system env vars can be used)

## Running in Different Environments

### Development (Default)
```bash
# Uses config.yaml and .env by default
go run ./cmd/zentube
```

### Explicit Development
```bash
export APP_ENV=development
go run ./cmd/zentube
```

### Staging
```bash
export APP_ENV=staging
go run ./cmd/zentube
```

### Production
```bash
export APP_ENV=production
./zentube
```

## Configuration Features

### Environment Detection

The application automatically detects the environment:
```go
env := config.GetEnvironment() // Returns Development, Staging, or Production
```

### Helper Methods

The Config struct provides helper methods:
```go
cfg.IsDevelopment() // true if APP_ENV=development
cfg.IsProduction()  // true if APP_ENV=production
cfg.IsStaging()     // true if APP_ENV=staging
```

### Validation

All configurations are validated at startup:
- App name must not be empty
- Port must be between 1 and 65535
- YouTube API key must be set
- Max results must be between 1 and 50
- Database path must be set

## Logging Configuration

Logging format changes based on environment:

- **Development**: Human-readable text format
  ```
  2024-01-15 10:30:45 INFO starting zentube env=development
  ```

- **Production/Staging**: JSON format for log aggregation
  ```json
  {"time":"2024-01-15T10:30:45Z","level":"INFO","msg":"starting zentube","env":"production"}
  ```

## Docker Deployment Example

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o zentube ./cmd/zentube

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/zentube .
COPY configs/config.production.yaml /root/configs/config.yaml

ENV APP_ENV=production
ENV YOUTUBE_API_KEY=your_key_here

EXPOSE 8080
CMD ["./zentube"]
```

## Best Practices

1. **Never commit `.env` files** - Add to `.gitignore`
2. **Use system environment variables in production** - Don't rely on .env files in containers
3. **Create environment templates** - Provide `.env.example` for developers
4. **Validate configs at startup** - Fail fast if configuration is invalid
5. **Use different API keys per environment** - Isolate development from production
6. **Set appropriate max_results** - Lower in dev to save quota, higher in prod for UX

## Environment Variable Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `APP_ENV` | No | development | Application environment (development, staging, production) |
| `YOUTUBE_API_KEY` | Yes | - | YouTube Data API v3 key |

## Troubleshooting

### Config file not found
- Ensure `configs/config.yaml` exists as fallback
- Check file paths are correct for your environment

### Database errors
- Verify database directory exists and is writable
- Production path `/var/lib/zentube/` requires proper permissions

### API key not loaded
- Check `.env` file exists and contains `YOUTUBE_API_KEY`
- Ensure environment-specific `.env.<env>` is loaded
- In production, verify system environment variable is set

## Implementation Details

The environment configuration system uses:
- `gopkg.in/yaml.v3` for YAML parsing
- `github.com/joho/godotenv` for .env file loading
- Standard library `os` package for environment variable access
- Custom validation logic for configuration integrity
