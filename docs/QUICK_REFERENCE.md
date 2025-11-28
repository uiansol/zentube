# Quick Reference: Implemented Patterns

Fast reference for developers looking up specific patterns in zentube.

## 🔥 High Priority Patterns

### 1. Structured Logging (slog)
**File:** `internal/adapters/http/middleware/slog.go`
```go
// Development: human-readable text
logger := middleware.NewLogger("development")

// Production: JSON for log aggregation
logger := middleware.NewLogger("production")

// Usage
logger.Info("request completed",
    slog.String("method", "POST"),
    slog.Int("status", 200),
    slog.Duration("duration", elapsed),
)
```

### 2. Configuration Validation
**File:** `internal/config/config.go`
```go
cfg, _ := config.LoadConfig("configs/config.yaml")
if err := cfg.Validate(); err != nil {
    log.Fatal(err) // Fail fast on invalid config
}
```

### 3. Health Check Endpoints
**File:** `internal/adapters/http/handlers/health_handler.go`
```bash
# Liveness probe (is server alive?)
curl http://localhost:8080/health/live

# Readiness probe (is database ready?)
curl http://localhost:8080/health/ready
```

### 4. Rate Limiting
**File:** `internal/adapters/http/middleware/ratelimit.go`
```go
// 10 requests per second, per IP
limiter := middleware.NewRateLimiter(10)
r.Use(limiter.Limit())
```

### 5. Security Headers
**File:** `internal/adapters/http/middleware/security.go`
```go
r.Use(middleware.SecureHeaders())
// Adds: X-Content-Type-Options, X-Frame-Options, etc.
```

### 6. HTTP Server Timeouts
**File:** `cmd/zentube/main.go`
```go
srv := &http.Server{
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       30 * time.Second,
    ReadHeaderTimeout: 5 * time.Second,
}
```

### 7. Panic Recovery
**File:** `internal/adapters/http/middleware/recovery.go`
```go
r.Use(middleware.Recovery(logger))
// Catches panics, logs stack trace, returns 500
```

### 8. Request ID Tracing
**File:** `internal/adapters/http/middleware/request_id.go`
```go
r.Use(middleware.RequestID())
// Adds X-Request-Id header to all requests/responses
// Access in handlers: c.GetString("request_id")
```

## 📊 Medium Priority Patterns

### 9. Custom Error Types
**File:** `internal/errors/errors.go`
```go
// Create domain errors
err := apperrors.NewValidationError("invalid input", "query too long")

// Automatically mapped to HTTP status
status := apperrors.GetStatusCode(err) // 400 for validation
```

### 10. Input Validation
**File:** `internal/validation/validation.go`
```go
validQuery, err := validation.ValidateSearchQuery(rawQuery)
if err != nil {
    // Returns ValidationError with details
}
// validQuery is sanitized and normalized
```

### 11. API Response Caching
**File:** `internal/cache/cache.go`, `internal/usecases/search_videos.go`
```go
// Create cache
cache := cache.NewCache(1000, 5*time.Minute)

// Check cache
if cached, found := cache.Get(key); found {
    return cached.([]Video), nil
}

// Store in cache
cache.Set(key, videos)
```

### 12. Environment-Specific Configuration
**File:** `internal/config/config.go`
```bash
# Development
APP_ENV=development go run ./cmd/zentube

# Staging
APP_ENV=staging go run ./cmd/zentube

# Production
APP_ENV=production ./zentube
```

## 🗄️ Database Patterns

### SQLite Production Setup
**File:** `internal/adapters/database/sqlite_repository.go`
```go
db, _ := sql.Open("sqlite3", "file:zentube.db?mode=rwc")

// WAL mode for concurrent reads
db.Exec("PRAGMA journal_mode=WAL")

// Connection pool
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// Prepared statements
stmt, _ := db.Prepare("INSERT INTO searches ...")
defer stmt.Close()
```

### Graceful Shutdown
**File:** `cmd/zentube/main.go`
```go
// Shutdown order is critical
srv.Shutdown(ctx)       // Stop accepting requests
stmt.Close()            // Close prepared statements
db.Close()              // Close database connections
```

## 🔧 Common Patterns

### Middleware Stack Order
```go
r.Use(middleware.Recovery(logger))      // 1. Catch panics first
r.Use(middleware.RequestID())           // 2. Add request ID
r.Use(middleware.Logging(logger))       // 3. Log with request ID
r.Use(middleware.SecureHeaders())       // 4. Add security headers
r.Use(middleware.NewRateLimiter(10).Limit()) // 5. Rate limit
```

### Error Response Format
```go
{
    "error": "validation failed",
    "details": "query exceeds maximum length",
    "status": 400
}
```

### Request Logging Format (JSON)
```json
{
    "time": "2024-01-15T10:30:45Z",
    "level": "INFO",
    "msg": "request completed",
    "request_id": "123e4567-e89b-12d3-a456-426614174000",
    "method": "GET",
    "path": "/search",
    "status": 200,
    "duration": "125ms",
    "ip": "192.168.1.1"
}
```

## 📂 Project Structure

```
zentube/
├── cmd/
│   └── zentube/
│       └── main.go                    # Application entry point
├── internal/
│   ├── adapters/
│   │   ├── database/
│   │   │   └── sqlite_repository.go   # SQLite implementation
│   │   ├── http/
│   │   │   ├── handlers/
│   │   │   │   ├── youtube_handler.go # YouTube search handler
│   │   │   │   ├── health_handler.go  # Health check handler
│   │   │   │   └── errors.go          # Error response handling
│   │   │   ├── middleware/
│   │   │   │   ├── slog.go           # Structured logging
│   │   │   │   ├── request_id.go     # Request ID generation
│   │   │   │   ├── security.go       # Security headers
│   │   │   │   ├── ratelimit.go      # Rate limiting
│   │   │   │   └── recovery.go       # Panic recovery
│   │   │   └── routes/
│   │   │       └── routes.go          # Route definitions
│   │   └── youtube/
│   │       └── youtube_client.go      # YouTube API client
│   ├── cache/
│   │   └── cache.go                   # In-memory TTL cache
│   ├── config/
│   │   └── config.go                  # Configuration loading
│   ├── entities/
│   │   ├── video.go                   # Video entity
│   │   └── search_history.go          # Search history entity
│   ├── errors/
│   │   └── errors.go                  # Custom error types
│   ├── ports/
│   │   ├── youtube_client.go          # YouTube port interface
│   │   └── search_history_repository.go # Repository port
│   ├── usecases/
│   │   ├── search_videos.go           # Search use case
│   │   └── search_videos_test.go      # Tests with mocks
│   └── validation/
│       └── validation.go               # Input validation
├── configs/
│   ├── config.yaml                    # Default config
│   ├── config.development.yaml        # Dev config
│   ├── config.staging.yaml            # Staging config
│   └── config.production.yaml         # Production config
├── docs/
│   ├── README.md                      # Full documentation index
│   ├── PRODUCTION_PATTERNS.md         # In-depth pattern guide
│   ├── DATABASE_PATTERNS.md           # Database optimization
│   ├── ENVIRONMENT_CONFIG.md          # Environment setup
│   └── QUICK_REFERENCE.md             # This file
└── web/
    ├── static/                        # CSS, JS assets
    └── templates/                     # Templ templates
```

## 🧪 Testing Examples

### Unit Test with Mocks
```go
func TestSearchVideos_Execute(t *testing.T) {
    mockClient := new(MockYouTubeClient)
    mockRepo := new(MockSearchHistoryRepository)
    
    expectedVideos := []entities.Video{{ID: "test123"}}
    mockClient.On("Search", "golang", int64(10)).Return(expectedVideos, nil)
    
    uc := usecases.NewSearchVideos(mockClient, mockRepo)
    videos, err := uc.Execute(ctx, "golang", 10)
    
    assert.NoError(t, err)
    assert.Len(t, videos, 1)
    mockClient.AssertExpectations(t)
}
```

### Cache Hit Test
```go
func TestCache_Hit(t *testing.T) {
    mockClient.On("Search", "golang", 10).Return(videos, nil).Once()
    
    // First call - cache miss
    result1, _ := uc.Execute(ctx, "golang", 10)
    
    // Second call - cache hit (no additional API call)
    result2, _ := uc.Execute(ctx, "golang", 10)
    
    mockClient.AssertNumberOfCalls(t, "Search", 1) // Only called once
}
```

## 🚀 Deployment Checklist

- [ ] Set `APP_ENV=production`
- [ ] Configure `YOUTUBE_API_KEY` in system env vars
- [ ] Use `config.production.yaml` with production database path
- [ ] Ensure database directory has write permissions
- [ ] Set up health check monitoring (`/health/ready`)
- [ ] Configure reverse proxy (nginx) for HTTPS
- [ ] Set appropriate rate limits for production traffic
- [ ] Enable structured JSON logging for log aggregation
- [ ] Review security headers for your domain
- [ ] Test graceful shutdown behavior

## 📖 Where to Learn More

- **Overview**: [docs/README.md](./README.md)
- **Pattern Details**: [docs/PRODUCTION_PATTERNS.md](./PRODUCTION_PATTERNS.md)
- **Database Guide**: [docs/DATABASE_PATTERNS.md](./DATABASE_PATTERNS.md)
- **Config Guide**: [docs/ENVIRONMENT_CONFIG.md](./ENVIRONMENT_CONFIG.md)

## 🔗 External References

- [Go slog Package](https://pkg.go.dev/log/slog)
- [Gin Framework](https://gin-gonic.com/)
- [SQLite WAL Mode](https://www.sqlite.org/wal.html)
- [12-Factor App](https://12factor.net/)
- [Rate Limiting (Token Bucket)](https://en.wikipedia.org/wiki/Token_bucket)
