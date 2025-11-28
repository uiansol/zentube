# Enterprise-Grade Production Patterns in Go

This document details the production-ready patterns implemented in zentube, focusing on pure Go solutions without external infrastructure dependencies.

## Table of Contents

1. [Structured Logging with slog](#1-structured-logging-with-slog)
2. [Configuration Validation](#2-configuration-validation)
3. [Health Check Endpoints](#3-health-check-endpoints)
4. [Rate Limiting](#4-rate-limiting)
5. [Security Headers](#5-security-headers)
6. [HTTP Server Timeouts](#6-http-server-timeouts)
7. [Enhanced Panic Recovery](#7-enhanced-panic-recovery)
8. [Request ID Tracing](#8-request-id-tracing)

---

## 1. Structured Logging with slog

### Implementation
**File:** `internal/adapters/http/middleware/slog.go`

### Why Structured Logging?

Traditional logging:
```go
log.Printf("[%s] %s - %d (%v)", method, path, statusCode, duration)
// Output: [POST] /search - 200 (234ms)
```

**Problems:**
- Hard to parse programmatically
- Difficult to filter by specific fields
- No log levels
- Can't be easily sent to log aggregation tools

Structured logging:
```go
logger.Info("request completed",
    slog.String("method", "POST"),
    slog.String("path", "/search"),
    slog.Int("status", 200),
    slog.Duration("duration", 234*time.Millisecond),
)
```

**Output (JSON):**
```json
{
  "time": "2025-11-27T10:30:45Z",
  "level": "INFO",
  "msg": "request completed",
  "method": "POST",
  "path": "/search",
  "status": 200,
  "duration": "234ms"
}
```

### Key Benefits

1. **Machine Parseable**: Easy integration with ELK, Datadog, CloudWatch
2. **Queryable**: Filter logs by any field
3. **Type Safe**: Compile-time checking of log fields
4. **Performance**: More efficient than string formatting
5. **Standard Library**: No external dependencies (Go 1.21+)

### Environment-Aware Logging

```go
func NewLogger(env string) *slog.Logger {
    if env == "production" {
        // JSON for machine parsing
        return slog.New(slog.NewJSONHandler(os.Stdout, opts))
    } else {
        // Human-readable for development
        return slog.New(slog.NewTextHandler(os.Stdout, opts))
    }
}
```

**Development output:**
```
time=2025-11-27T10:30:45 level=INFO msg="request completed" method=POST path=/search
```

**Production output:**
```json
{"time":"2025-11-27T10:30:45Z","level":"INFO","msg":"request completed","method":"POST"}
```

### Log Levels

```go
slog.Debug("detailed info", ...)   // Development debugging
slog.Info("normal operation", ...) // Default level
slog.Warn("warning condition", ...) // Potential issues
slog.Error("error occurred", ...)  // Errors requiring attention
```

### Best Practices

✅ **DO:**
- Use consistent field names across logs (`request_id`, not `requestId` sometimes and `req_id` other times)
- Include request_id in all logs for tracing
- Log at appropriate levels
- Use context-aware logging when possible

❌ **DON'T:**
- Log sensitive information (passwords, API keys, tokens)
- Log at DEBUG level in production (performance impact)
- Use string interpolation in log messages
- Forget to include error context

---

## 2. Configuration Validation

### Implementation
**File:** `internal/config/config.go` - `Validate()` method

### The Problem

Without validation, misconfigurations are discovered at runtime:

```go
// Application starts successfully but fails later
port := cfg.App.Port  // 0 (invalid!)
srv := &http.Server{Addr: ":0"}  // Binds to random port!
```

### The Solution

**Fail Fast at Startup:**

```go
func (c *Config) Validate() error {
    var errs []error
    
    // Validate each field
    if c.App.Port < 1 || c.App.Port > 65535 {
        errs = append(errs, fmt.Errorf("app.port must be between 1 and 65535, got %d", c.App.Port))
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("configuration validation failed: %v", errs)
    }
    
    return nil
}
```

### Validation Categories

#### 1. **Required Fields**
```go
if c.YouTube.APIKey == "" {
    errs = append(errs, errors.New("youtube.api_key cannot be empty"))
}
```

#### 2. **Range Validation**
```go
if c.YouTube.MaxResults < 1 || c.YouTube.MaxResults > 50 {
    errs = append(errs, fmt.Errorf("youtube.max_results must be between 1 and 50, got %d", c.YouTube.MaxResults))
}
```

#### 3. **Format Validation**
```go
if !strings.HasSuffix(c.Database.Path, ".db") {
    errs = append(errs, errors.New("database.path must end with .db"))
}
```

#### 4. **Semantic Validation**
```go
if c.App.ReadTimeout > c.App.WriteTimeout {
    errs = append(errs, errors.New("read_timeout should not exceed write_timeout"))
}
```

### Benefits

- **Clear Error Messages**: Know exactly what's wrong
- **Fast Feedback**: Fail before processing any requests
- **Self-Documenting**: Validation rules document expected values
- **Testing**: Easier to test configuration scenarios

### Testing Configuration

```go
func TestConfig_Validate_InvalidPort(t *testing.T) {
    cfg := &Config{
        App: App{Port: 99999}, // Invalid
    }
    
    err := cfg.Validate()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "port must be between")
}
```

---

## 3. Health Check Endpoints

### Implementation
**File:** `internal/adapters/http/handlers/health_handler.go`

### Two Types of Health Checks

#### Liveness Probe (`/health/live`)
**Question:** "Is the server running?"

```go
func (h *HealthHandler) Live(c *gin.Context) {
    c.JSON(http.StatusOK, HealthResponse{
        Status:    "ok",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    })
}
```

**Use Cases:**
- Kubernetes liveness probe
- Load balancer health check
- Simple uptime monitoring

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2025-11-27T10:30:45Z"
}
```

#### Readiness Probe (`/health/ready`)
**Question:** "Is the server ready to handle traffic?"

```go
func (h *HealthHandler) Ready(c *gin.Context) {
    checks := make(map[string]string)
    allHealthy := true
    
    // Check database
    ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
    defer cancel()
    
    if err := h.db.PingContext(ctx); err != nil {
        checks["database"] = "unhealthy: " + err.Error()
        allHealthy = false
    } else {
        checks["database"] = "healthy"
    }
    
    if !allHealthy {
        return http.StatusServiceUnavailable
    }
    
    c.JSON(http.StatusOK, HealthResponse{...})
}
```

**Healthy Response (200):**
```json
{
  "status": "ok",
  "timestamp": "2025-11-27T10:30:45Z",
  "checks": {
    "database": "healthy"
  }
}
```

**Unhealthy Response (503):**
```json
{
  "status": "degraded",
  "timestamp": "2025-11-27T10:30:45Z",
  "checks": {
    "database": "unhealthy: connection timeout"
  }
}
```

### Orchestration Integration

#### Docker Compose
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health/live"]
  interval: 30s
  timeout: 10s
  retries: 3
```

#### Kubernetes
```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

### What to Check in Readiness

✅ **Should Check:**
- Database connectivity
- Required external APIs
- Critical file system access
- Cache connectivity (if required)

❌ **Should NOT Check:**
- Optional features
- Nice-to-have integrations
- Slow operations (>2 seconds)

### Timeout Considerations

Always use context with timeout:
```go
ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
defer cancel()

if err := h.db.PingContext(ctx); err != nil {
    // Mark as unhealthy
}
```

**Why?** Prevents health check from hanging and causing cascading failures.

---

## 4. Rate Limiting

### Implementation
**File:** `internal/adapters/http/middleware/ratelimit.go`

### The Problem

Without rate limiting:
- Vulnerable to DoS attacks
- API quota exhaustion (YouTube API)
- Resource starvation
- Unfair resource distribution

### Token Bucket Algorithm

```go
limiter := rate.NewLimiter(rate.Limit(10), 20)
// 10 tokens/second
// Burst capacity: 20 tokens
```

**How it works:**
1. Bucket holds max 20 tokens
2. Refills at 10 tokens/second
3. Each request consumes 1 token
4. If bucket empty, request denied

**Example:**
```
Time 0s:  20 tokens (full)
         ↓ 20 requests ✓ (burst handled)
Time 0s:  0 tokens
         ↓ 1 request  ✗ (rate limited)
Time 0.1s: 1 token (refilled)
         ↓ 1 request  ✓
```

### Per-IP Rate Limiting

```go
type IPRateLimiter struct {
    ips map[string]*rate.Limiter  // One limiter per IP
    mu  sync.RWMutex               // Thread-safe access
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()
    
    limiter, exists := i.ips[ip]
    if !exists {
        limiter = rate.NewLimiter(i.r, i.b)
        i.ips[ip] = limiter
    }
    
    return limiter
}
```

**Benefits:**
- One user can't block others
- Fair resource distribution
- Isolated abuse containment

### Configuration

```go
// 10 requests/second, burst of 20
RateLimit(rate.Limit(10), 20, logger)
```

**Choosing values:**

**Requests per second (r):**
- High traffic app: 100-1000
- Medium traffic app: 10-100
- Low traffic app: 1-10

**Burst (b):**
- Usually 2x of rate
- Handles legitimate bursts
- Too high = defeats purpose
- Too low = frustrates users

### Response Headers

```go
c.Header("X-RateLimit-Limit", "10")       // Max requests per period
c.Header("X-RateLimit-Remaining", "0")    // Requests left
c.Header("Retry-After", "60")             // Wait time in seconds
```

**Client can:**
- Display countdown timer
- Implement exponential backoff
- Inform user of limitations

### Memory Management

**Problem:** Map grows indefinitely as new IPs connect

**Solutions:**

1. **TTL-based cleanup** (recommended):
```go
// Remove limiters not used in 1 hour
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        i.mu.Lock()
        for ip, limiter := range i.ips {
            if limiter.lastUsed.Before(time.Now().Add(-1 * time.Hour)) {
                delete(i.ips, ip)
            }
        }
        i.mu.Unlock()
    }
}()
```

2. **LRU Cache**: Limit to N most recent IPs

3. **External Store**: Redis with TTL

### Testing Rate Limits

```go
func TestRateLimit(t *testing.T) {
    limiter := NewIPRateLimiter(rate.Limit(2), 2, logger)
    
    // First 2 requests should succeed
    assert.True(t, limiter.GetLimiter("1.2.3.4").Allow())
    assert.True(t, limiter.GetLimiter("1.2.3.4").Allow())
    
    // Third should fail
    assert.False(t, limiter.GetLimiter("1.2.3.4").Allow())
}
```

---

## 5. Security Headers

### Implementation
**File:** `internal/adapters/http/middleware/security.go`

### Why Security Headers Matter

Headers are your **first line of defense** against common web vulnerabilities.

### Headers Implemented

#### 1. X-Content-Type-Options

```go
c.Header("X-Content-Type-Options", "nosniff")
```

**Prevents:** MIME type sniffing attacks

**Example Attack:**
```html
<!-- Attacker uploads image.jpg -->
<script src="/uploads/image.jpg"></script>
<!-- Without nosniff, browser might execute it as JavaScript! -->
```

**With nosniff:** Browser respects Content-Type, won't execute images as scripts.

#### 2. X-Frame-Options

```go
c.Header("X-Frame-Options", "DENY")
```

**Prevents:** Clickjacking attacks

**Example Attack:**
```html
<iframe src="https://yoursite.com/transfer-money"></iframe>
<!-- Invisible iframe overlaid on fake button -->
<!-- User thinks they're clicking "Play Game" but actually click "Transfer $1000" -->
```

**With DENY:** Page cannot be embedded in iframes.

**Alternatives:**
- `SAMEORIGIN`: Allow same domain iframes
- `ALLOW-FROM https://trusted.com`: Allow specific domains

#### 3. X-XSS-Protection

```go
c.Header("X-XSS-Protection", "1; mode=block")
```

**Prevents:** Cross-Site Scripting (XSS) in legacy browsers

**Note:** Modern browsers use CSP instead, but this provides defense-in-depth.

#### 4. Content-Security-Policy (CSP)

```go
c.Header("Content-Security-Policy", 
    "default-src 'self'; " +
    "img-src 'self' https://i.ytimg.com; " +
    "script-src 'self' 'unsafe-inline'; " +
    "style-src 'self' 'unsafe-inline'")
```

**Most powerful security header.** Controls what resources can load.

**Directives:**
- `default-src 'self'`: Only load from same origin
- `img-src`: Allowed image sources
- `script-src`: Allowed script sources
- `style-src`: Allowed stylesheet sources

**Example Prevention:**
```html
<!-- Attacker injects: -->
<script src="https://evil.com/steal-cookies.js"></script>

<!-- CSP blocks: "https://evil.com not in script-src whitelist" -->
```

**'unsafe-inline' note:** Allows inline scripts/styles. Better to avoid:
```go
// Instead of: <script>alert('hi')</script>
// Use: <script src="/js/app.js"></script>
```

#### 5. Referrer-Policy

```go
c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
```

**Controls:** What referrer information is sent

**Policies:**
- `strict-origin-when-cross-origin`: Send full URL to same origin, only origin to others
- `no-referrer`: Never send referrer
- `same-origin`: Only send to same origin

**Why it matters:**
```
User on: https://yourapp.com/user/john/private-documents
Clicks link to: https://external-site.com

Without policy: External site sees full URL with "private-documents"
With strict-origin: External site only sees "https://yourapp.com"
```

#### 6. Permissions-Policy

```go
c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
```

**Controls:** Browser feature access

**Prevents:**
- Malicious scripts accessing camera/microphone
- Location tracking without consent
- Battery draining from unused features

**Syntax:**
- `feature=()`: Deny all
- `feature=(self)`: Allow same origin
- `feature=(self "https://trusted.com")`: Allow specific domains

#### 7. Strict-Transport-Security (HSTS)

```go
func HSTS(maxAge int) gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
            c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        }
        c.Next()
    }
}
```

**Forces:** HTTPS for specified duration

**Example:**
```
User types: http://yoursite.com
Browser automatically converts to: https://yoursite.com
```

**Parameters:**
- `max-age=31536000`: 1 year in seconds
- `includeSubDomains`: Apply to all subdomains
- `preload`: Submit to browser preload list

**Important:** Only use with HTTPS properly configured!

### Testing Security Headers

```bash
curl -I https://yoursite.com

HTTP/1.1 200 OK
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
```

**Online tools:**
- [securityheaders.com](https://securityheaders.com)
- [Mozilla Observatory](https://observatory.mozilla.org)

---

## 6. HTTP Server Timeouts

### Implementation
**File:** `cmd/zentube/main.go`

### The Problem

Without timeouts, a single slow client can exhaust server resources:

```go
// Default (no timeouts)
srv := &http.Server{
    Addr:    ":8080",
    Handler: r,
}
// ❌ Vulnerable to Slowloris attack
// ❌ Hanging connections consume memory
// ❌ Can lead to resource exhaustion
```

### The Solution

```go
srv := &http.Server{
    Addr:           ":8080",
    Handler:        r,
    ReadTimeout:    10 * time.Second,   // ✅
    WriteTimeout:   10 * time.Second,   // ✅
    IdleTimeout:    60 * time.Second,   // ✅
    MaxHeaderBytes: 1 << 20,            // ✅ 1 MB
}
```

### Timeout Types

#### 1. ReadTimeout

**What it covers:**
- Reading request headers
- Reading request body

**When it triggers:**
```
Client connects
 ↓
Starts sending headers slowly...
 ↓
10 seconds pass
 ↓
Server closes connection ✂️
```

**Prevents:**
- Slowloris attacks (slow headers)
- Slow body DoS attacks
- Half-open connections

**Choosing value:**
```go
ReadTimeout: 10 * time.Second  // Most APIs
ReadTimeout: 30 * time.Second  // File upload endpoints
ReadTimeout: 5 * time.Second   // Health checks
```

#### 2. WriteTimeout

**What it covers:**
- Writing response headers
- Writing response body

**When it triggers:**
```
Server generates response
 ↓
Starts sending to slow client...
 ↓
10 seconds pass
 ↓
Server closes connection ✂️
```

**Prevents:**
- Slow-read attacks (client reads slowly)
- Resources locked for slow clients
- Memory exhaustion from buffered responses

**Choosing value:**
```go
WriteTimeout: 10 * time.Second   // Most APIs
WriteTimeout: 60 * time.Second   // Large responses
WriteTimeout: 120 * time.Second  // File downloads
```

#### 3. IdleTimeout

**What it covers:**
- Time between requests on keep-alive connections

**When it triggers:**
```
Request 1 completes
 ↓
Connection kept alive (HTTP Keep-Alive)
 ↓
60 seconds pass with no new request
 ↓
Server closes connection ✂️
```

**Prevents:**
- Idle connections consuming memory
- Connection pool exhaustion
- Port exhaustion

**Choosing value:**
```go
IdleTimeout: 60 * time.Second   // Standard
IdleTimeout: 120 * time.Second  // High-latency clients
IdleTimeout: 30 * time.Second   // High-traffic servers
```

#### 4. MaxHeaderBytes

**What it covers:**
- Maximum size of request headers

```go
MaxHeaderBytes: 1 << 20  // 1 MB
```

**Prevents:**
- Memory exhaustion from huge headers
- DoS via large Cookie headers
- Buffer overflow attacks

**Example attack:**
```http
GET / HTTP/1.1
Host: example.com
Cookie: [10 MB of data...]  ← Rejected!
```

### Timeout Interactions

**Important:** ReadTimeout includes time to read headers AND body

```go
ReadTimeout: 10 * time.Second

// Upload 100 MB file
// Transfer rate: 1 MB/s
// Required time: 100 seconds
// Will timeout! ⚠️

// Solution: Longer ReadTimeout for upload endpoints
```

### Context Timeouts

Server timeouts are **last resort**. Better to use context:

```go
// Handler
func (h *Handler) Search(c *gin.Context) {
    // Add timeout to specific operation
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()
    
    videos, err := h.searchUC.Execute(ctx, query, maxResults)
    // Operation will abort after 5 seconds
}
```

**Benefits:**
- Granular control
- Different timeouts for different operations
- Graceful cancellation

### Testing Timeouts

```go
func TestReadTimeout(t *testing.T) {
    srv := &http.Server{
        Addr:        ":8080",
        ReadTimeout: 1 * time.Second,
    }
    
    // Send request very slowly
    conn, _ := net.Dial("tcp", "localhost:8080")
    time.Sleep(2 * time.Second)
    conn.Write([]byte("GET / HTTP/1.1\r\n"))
    
    // Should be closed
    assert.Error(t, conn.Read(buf))
}
```

---

## 7. Enhanced Panic Recovery

### Implementation
**File:** `internal/adapters/http/middleware/recovery.go`

### Why Custom Recovery?

Gin's default recovery:
```go
// Catches panic but limited logging
gin.Default() // Includes recovery middleware
```

**Problems:**
- Minimal context in logs
- No stack trace in structured logs
- Missing request details
- Generic error to client

### Our Implementation

```go
func Recovery(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // Get request ID
                requestID, _ := c.Get("request_id")
                
                // Log with full context
                logger.Error("panic recovered",
                    slog.Any("error", err),
                    slog.String("request_id", requestID.(string)),
                    slog.String("method", c.Request.Method),
                    slog.String("path", c.Request.URL.Path),
                    slog.String("client_ip", c.ClientIP()),
                    slog.String("stack", string(debug.Stack())),
                )
                
                // Return user-friendly error
                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                    "error": "internal server error",
                    "message": "An unexpected error occurred. Please try again later.",
                    "request_id": requestID,
                })
            }
        }()
        c.Next()
    }
}
```

### What We Log

#### 1. The Error
```go
slog.Any("error", err)
```
Could be:
- `runtime error: invalid memory address`
- `runtime error: index out of range`
- Custom panic: `panic("database connection lost")`

#### 2. Request ID
```go
slog.String("request_id", requestID.(string))
```
**Enables:** Trace this panic across all related logs

#### 3. Request Details
```go
slog.String("method", c.Request.Method),
slog.String("path", c.Request.URL.Path),
slog.String("client_ip", c.ClientIP()),
```
**Answers:** What caused this panic?

#### 4. Stack Trace
```go
slog.String("stack", string(debug.Stack()))
```

**Example output:**
```
goroutine 42 [running]:
runtime/debug.Stack()
    /usr/local/go/src/runtime/debug/stack.go:24 +0x65
github.com/yourapp/middleware.Recovery.func1.1()
    /app/middleware/recovery.go:18 +0x125
github.com/yourapp/handlers.(*Handler).Search()
    /app/handlers/youtube.go:45 +0x234
```

**Enables:** Exact line number where panic occurred

### Response to Client

```json
{
  "error": "internal server error",
  "message": "An unexpected error occurred. Please try again later.",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Why generic?**
- Don't expose internal details
- Security best practice
- Provide request_id for support

**Customer support:**
> "I got an error!"  
> "Please provide the request_id shown in the error"  
> *Searches logs by request_id*  
> *Finds exact panic with stack trace*

### Common Panic Causes

#### 1. Nil Pointer
```go
var user *User
fmt.Println(user.Name)  // panic: invalid memory address
```

#### 2. Index Out of Bounds
```go
videos := []Video{}
video := videos[0]  // panic: index out of range
```

#### 3. Type Assertion
```go
value := c.Get("user")
user := value.(User)  // panic if value is not User
```

#### 4. Channel Closed
```go
ch := make(chan int)
close(ch)
ch <- 1  // panic: send on closed channel
```

### Prevention > Recovery

**Best practices:**
```go
// Check for nil
if user != nil {
    fmt.Println(user.Name)
}

// Check slice bounds
if len(videos) > 0 {
    video := videos[0]
}

// Safe type assertion
if user, ok := value.(User); ok {
    // Use user
}

// Don't close channels you don't own
// Or use sync patterns
```

### Monitoring

Track panic rate:
```go
var panicCounter int64

defer func() {
    if err := recover(); err != nil {
        atomic.AddInt64(&panicCounter, 1)
        // Log...
    }
}()
```

**Alert if:**
- Panic rate > 1% of requests
- Sudden spike in panics
- Same panic repeatedly

---

## 8. Request ID Tracing

### Implementation
**File:** `internal/adapters/http/middleware/request_id.go`

### The Problem

**Debugging without request IDs:**

```
2025-11-27 10:30:45 INFO request started path=/search
2025-11-27 10:30:45 INFO request started path=/health
2025-11-27 10:30:45 ERROR database timeout
2025-11-27 10:30:46 INFO request started path=/search
2025-11-27 10:30:46 INFO request completed path=/health
2025-11-27 10:30:47 ERROR youtube api failed
2025-11-27 10:30:47 INFO request completed path=/search
```

**Question:** Which search request failed?  
**Answer:** 🤷 Can't tell!

### The Solution

**With request IDs:**

```
2025-11-27 10:30:45 INFO request started path=/search request_id=abc-123
2025-11-27 10:30:45 INFO request started path=/health request_id=def-456
2025-11-27 10:30:45 ERROR database timeout request_id=abc-123
2025-11-27 10:30:46 INFO request started path=/search request_id=ghi-789
2025-11-27 10:30:46 INFO request completed path=/health request_id=def-456
2025-11-27 10:30:47 ERROR youtube api failed request_id=abc-123
2025-11-27 10:30:47 INFO request completed path=/search request_id=abc-123
```

**Filter by `request_id=abc-123`:**
```
10:30:45 INFO request started path=/search
10:30:45 ERROR database timeout
10:30:47 ERROR youtube api failed
10:30:47 INFO request completed path=/search
```

**Now we know:** First search request had both DB timeout and YouTube API failure!

### Implementation

```go
func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check if load balancer already set one
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            // Generate new UUID
            requestID = uuid.New().String()
        }
        
        // Store in context for handlers
        c.Set("request_id", requestID)
        
        // Return in response
        c.Header("X-Request-ID", requestID)
        
        c.Next()
    }
}
```

### UUID Format

```go
uuid.New().String()
// Output: "550e8400-e29b-41d4-a716-446655440000"
```

**Properties:**
- Globally unique
- 128-bit number
- Standardized format (RFC 4122)
- Collision probability: ~0%

**Alternatives:**
- `ulid`: Sortable by time
- `ksuid`: Time-sortable, shorter
- Random string: Custom format

### Integration Points

#### 1. Middleware
```go
func Middleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID, _ := c.Get("request_id")
        
        logger.Info("request completed",
            slog.String("request_id", requestID.(string)),
            // ... other fields
        )
    }
}
```

#### 2. Error Handler
```go
func respondError(c *gin.Context, statusCode int, message string) {
    requestID, _ := c.Get("request_id")
    
    slog.Error("request error",
        slog.String("request_id", requestID.(string)),
        // ... other fields
    )
}
```

#### 3. Response
```http
HTTP/1.1 200 OK
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
Content-Type: application/json

{"data": [...]}
```

**Client can:**
- Include in bug reports
- Reference in support tickets
- Use for retry logic

### Propagation

**Microservices:**
```go
func callOtherService(ctx context.Context) {
    requestID, _ := ctx.Value("request_id").(string)
    
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    req.Header.Set("X-Request-ID", requestID)  // Propagate!
    
    resp, _ := http.DefaultClient.Do(req)
}
```

**Distributed tracing:**
```
Service A [request_id=abc-123]
    ↓
Service B [request_id=abc-123]
    ↓
Service C [request_id=abc-123]
```

All services log with same request_id → Full trace!

### Log Aggregation

**ELK Stack:**
```json
{
  "query": {
    "term": {
      "request_id": "550e8400-e29b-41d4-a716-446655440000"
    }
  }
}
```

**CloudWatch Insights:**
```
fields @timestamp, message, request_id
| filter request_id = "550e8400-e29b-41d4-a716-446655440000"
| sort @timestamp asc
```

**Datadog:**
```
request_id:550e8400-e29b-41d4-a716-446655440000
```

### Best Practices

✅ **DO:**
- Generate at edge (API gateway/first service)
- Propagate to all downstream services
- Include in all logs
- Return to client
- Use standard header name (`X-Request-ID`)

❌ **DON'T:**
- Change request_id mid-request
- Skip logging request_id
- Generate new ID for each service
- Use sequential numbers (security risk)

---

## Summary

These eight patterns form the foundation of production-ready Go applications:

1. **Structured Logging**: Machine-parseable, queryable logs
2. **Config Validation**: Fail fast with clear errors
3. **Health Checks**: Enable orchestration and monitoring
4. **Rate Limiting**: Protect against abuse and resource exhaustion
5. **Security Headers**: Defense against common web attacks
6. **HTTP Timeouts**: Prevent resource leaks and DoS
7. **Panic Recovery**: Graceful handling with full context
8. **Request ID**: End-to-end request tracing

**All implemented with pure Go** - no external infrastructure required.

---

## Medium Priority Patterns

### 9. Custom Error Types

### Implementation
**Files:** 
- `internal/errors/errors.go` - Error type definitions
- `internal/adapters/http/handlers/errors.go` - Error response handling

### Why Custom Error Types?

Generic errors:
```go
return nil, errors.New("invalid query")
// How should HTTP handler map this to status code?
```

**Problems:**
- No context about error category
- Can't differentiate client vs server errors
- Loss of information across boundaries
- Inconsistent HTTP status codes

Custom error types:
```go
return nil, apperrors.NewValidationError("invalid query", "query length exceeds maximum")
```

### Pattern: Domain-Specific Errors

```go
// internal/errors/errors.go
package errors

import "net/http"

// ErrorType categorizes errors for appropriate handling
type ErrorType string

const (
    ErrorTypeValidation  ErrorType = "VALIDATION_ERROR"
    ErrorTypeNotFound    ErrorType = "NOT_FOUND"
    ErrorTypeInternal    ErrorType = "INTERNAL_ERROR"
    ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"
)

// AppError represents application-specific errors with HTTP context
type AppError struct {
    Type    ErrorType
    Message string
    Details string
    Err     error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return e.Message + ": " + e.Err.Error()
    }
    return e.Message
}

// Unwrap enables error chain inspection
func (e *AppError) Unwrap() error {
    return e.Err
}

// Constructor functions for common error types
func NewValidationError(message, details string) *AppError {
    return &AppError{
        Type:    ErrorTypeValidation,
        Message: message,
        Details: details,
    }
}

func NewNotFoundError(message string) *AppError {
    return &AppError{
        Type:    ErrorTypeNotFound,
        Message: message,
    }
}

func NewInternalError(message string, err error) *AppError {
    return &AppError{
        Type:    ErrorTypeInternal,
        Message: message,
        Err:     err,
    }
}

// GetStatusCode maps error types to HTTP status codes
func GetStatusCode(err error) int {
    if appErr, ok := err.(*AppError); ok {
        switch appErr.Type {
        case ErrorTypeValidation:
            return http.StatusBadRequest
        case ErrorTypeNotFound:
            return http.StatusNotFound
        case ErrorTypeUnauthorized:
            return http.StatusUnauthorized
        default:
            return http.StatusInternalServerError
        }
    }
    return http.StatusInternalServerError
}
```

### Usage in Handlers

```go
// Before: Generic error handling
func (h *Handler) Search(c *gin.Context) {
    videos, err := h.useCase.Execute(ctx, query, maxResults)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()}) // Always 500!
        return
    }
}

// After: Type-aware error handling
func (h *Handler) Search(c *gin.Context) {
    videos, err := h.useCase.Execute(ctx, query, maxResults)
    if err != nil {
        respondError(c, err)
        return
    }
}

func respondError(c *gin.Context, err error) {
    statusCode := apperrors.GetStatusCode(err)
    
    response := ErrorResponse{
        Error:   err.Error(),
        Status:  statusCode,
    }
    
    // Add details for non-500 errors
    if appErr, ok := err.(*apperrors.AppError); ok && statusCode < 500 {
        response.Details = appErr.Details
    }
    
    c.JSON(statusCode, response)
}
```

### Benefits

1. **Consistent Status Codes**: Automatic mapping of error types to HTTP codes
2. **Error Context**: Details field provides additional information
3. **Error Wrapping**: Standard `errors.Unwrap` support
4. **Type Safety**: Strongly typed error categories
5. **Separation of Concerns**: Business logic doesn't need to know about HTTP

### Production Gotchas

- **Don't expose internal errors in production**: Hide stack traces and internal details in 500 responses
- **Use details carefully**: Only expose validation details to clients, not internal errors
- **Log everything**: Even client errors should be logged for analysis
- **Maintain error chain**: Use `Unwrap()` for error inspection

---

### 10. Input Validation

### Implementation
**File:** `internal/validation/validation.go`

### Why Input Validation?

Without validation:
```go
func (h *Handler) Search(c *gin.Context) {
    query := c.Query("q")
    // What if query is empty, too long, or contains SQL injection?
    videos, err := h.useCase.Execute(ctx, query, maxResults)
}
```

**Security Risks:**
- SQL injection (if not using parameterized queries)
- XSS attacks
- Resource exhaustion (huge queries)
- Control character attacks

### Pattern: Centralized Validation

```go
// internal/validation/validation.go
package validation

import (
    "strings"
    "unicode"
    apperrors "github.com/uiansol/zentube/internal/errors"
)

const (
    MaxQueryLength = 200
    MinQueryLength = 1
)

// ValidateSearchQuery validates and sanitizes search queries
func ValidateSearchQuery(query string) (string, error) {
    // Check empty
    if strings.TrimSpace(query) == "" {
        return "", apperrors.NewValidationError(
            "search query cannot be empty",
            "please provide a search term",
        )
    }
    
    // Check length
    if len(query) > MaxQueryLength {
        return "", apperrors.NewValidationError(
            "search query too long",
            fmt.Sprintf("maximum length is %d characters", MaxQueryLength),
        )
    }
    
    // Sanitize: remove control characters and normalize whitespace
    sanitized := sanitizeString(query)
    normalized := normalizeWhitespace(sanitized)
    
    return normalized, nil
}

// sanitizeString removes potentially dangerous control characters
func sanitizeString(s string) string {
    return strings.Map(func(r rune) rune {
        // Keep printable characters, spaces, and common punctuation
        if unicode.IsPrint(r) || unicode.IsSpace(r) {
            return r
        }
        return -1 // Remove control characters
    }, s)
}

// normalizeWhitespace replaces multiple spaces with single space
func normalizeWhitespace(s string) string {
    return strings.Join(strings.Fields(s), " ")
}
```

### Usage in Handler

```go
func (h *YouTubeHandler) Search(c *gin.Context) {
    query := c.Query("q")
    
    // Validate and sanitize
    validatedQuery, err := validation.ValidateSearchQuery(query)
    if err != nil {
        respondError(c, err)
        return
    }
    
    // Now safe to use
    videos, err := h.searchVideos.Execute(c.Request.Context(), validatedQuery, h.maxResults)
}
```

### Validation Checklist

**String Validation:**
- [ ] Length limits (prevent resource exhaustion)
- [ ] Character whitelist (prevent injection)
- [ ] Sanitization (remove dangerous characters)
- [ ] Normalization (consistent format)

**Numeric Validation:**
- [ ] Range checks (min/max values)
- [ ] Overflow protection
- [ ] Type validation (int vs float)

**General:**
- [ ] Required fields
- [ ] Format validation (email, URL, etc.)
- [ ] Business logic validation

### Benefits

1. **Security**: Prevents injection attacks and XSS
2. **Data Quality**: Ensures consistent, clean data
3. **User Experience**: Clear error messages
4. **Resource Protection**: Length limits prevent abuse
5. **Maintainability**: Centralized validation logic

### Production Gotchas

- **Validate at boundaries**: API handlers, not just use cases
- **Fail fast**: Validate before expensive operations
- **Don't trust clients**: Even internal clients can send bad data
- **Balance strictness**: Too strict = bad UX, too loose = security risk

---

### 11. API Response Caching

### Implementation
**Files:**
- `internal/cache/cache.go` - In-memory TTL cache
- `internal/usecases/search_videos.go` - Cache integration

### Why Caching?

Without caching:
```go
func (s *SearchVideos) Execute(query string) ([]Video, error) {
    return s.ytClient.Search(query) // Every request = API call
}
```

**Problems:**
- API quota exhaustion (YouTube allows limited requests/day)
- Slow response times (external API latency)
- Cost (some APIs charge per request)
- Unnecessary load on external services

### Pattern: In-Memory TTL Cache

```go
// internal/cache/cache.go
package cache

import (
    "sync"
    "time"
)

type Cache struct {
    mu         sync.RWMutex
    items      map[string]*cacheItem
    maxEntries int
    ttl        time.Duration
}

type cacheItem struct {
    value      interface{}
    expiration time.Time
}

func NewCache(maxEntries int, ttl time.Duration) *Cache {
    c := &Cache{
        items:      make(map[string]*cacheItem),
        maxEntries: maxEntries,
        ttl:        ttl,
    }
    
    // Start cleanup goroutine
    go c.cleanupExpired()
    
    return c
}

// Get retrieves value from cache
func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    item, found := c.items[key]
    if !found {
        return nil, false
    }
    
    // Check expiration
    if time.Now().After(item.expiration) {
        return nil, false
    }
    
    return item.value, true
}

// Set stores value in cache
func (c *Cache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // Evict old entries if cache is full
    if len(c.items) >= c.maxEntries {
        c.evictOldest()
    }
    
    c.items[key] = &cacheItem{
        value:      value,
        expiration: time.Now().Add(c.ttl),
    }
}

// cleanupExpired removes expired entries periodically
func (c *Cache) cleanupExpired() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        c.mu.Lock()
        now := time.Now()
        for k, v := range c.items {
            if now.After(v.expiration) {
                delete(c.items, k)
            }
        }
        c.mu.Unlock()
    }
}

// evictOldest removes the oldest entry (FIFO)
func (c *Cache) evictOldest() {
    var oldestKey string
    var oldestTime time.Time
    
    for k, v := range c.items {
        if oldestTime.IsZero() || v.expiration.Before(oldestTime) {
            oldestKey = k
            oldestTime = v.expiration
        }
    }
    
    if oldestKey != "" {
        delete(c.items, oldestKey)
    }
}
```

### Cache Key Generation

```go
// GenerateKey creates consistent cache keys
func GenerateKey(prefix string, parts ...interface{}) string {
    key := prefix
    for _, part := range parts {
        key += fmt.Sprintf(":%v", part)
    }
    return key
}

// Usage:
cacheKey := cache.GenerateKey("search", query, maxResults)
// Result: "search:golang:10"
```

### Integration in Use Case

```go
func (s *SearchVideos) Execute(ctx context.Context, query string, maxResults int64) ([]Video, error) {
    // Generate cache key
    cacheKey := cache.GenerateKey("search", query, maxResults)
    
    // Try cache first
    if s.cache != nil {
        if cached, found := s.cache.Get(cacheKey); found {
            if videos, ok := cached.([]entities.Video); ok {
                return videos, nil // Cache hit!
            }
        }
    }
    
    // Cache miss - fetch from API
    videos, err := s.ytClient.Search(query, maxResults)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    if s.cache != nil {
        s.cache.Set(cacheKey, videos)
    }
    
    return videos, nil
}
```

### Cache Statistics

```go
// GetStats returns cache metrics
func (c *Cache) GetStats() map[string]interface{} {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    return map[string]interface{}{
        "total_entries": len(c.items),
        "max_entries":   c.maxEntries,
        "ttl_seconds":   c.ttl.Seconds(),
    }
}
```

### Benefits

1. **Reduced API Costs**: Fewer external API calls
2. **Lower Latency**: In-memory access is microseconds vs API milliseconds
3. **Quota Protection**: Stay within API rate limits
4. **Resilience**: Cache can serve requests if API is temporarily down
5. **Thread-Safe**: RWMutex allows concurrent reads

### Cache Design Decisions

**TTL Selection:**
- Too short: Cache doesn't help much
- Too long: Stale data
- **Zentube: 5 minutes** - Good balance for search results

**Eviction Policy:**
- **FIFO** (First-In-First-Out): Simple, predictable
- LRU (Least Recently Used): Better hit rate but more complex
- LFU (Least Frequently Used): Best hit rate, most complex

**Size Limits:**
- **1000 entries** in zentube
- Each video ~1KB → ~1MB total memory
- Prevents unbounded memory growth

### Production Gotchas

- **Memory limits**: Monitor cache size in production
- **Cache invalidation**: "There are only two hard things in Computer Science: cache invalidation and naming things"
- **Thundering herd**: Many requests for same uncached item can overwhelm API
- **Serialization**: In-memory cache is lost on restart
- **Distributed systems**: This cache is per-instance, not shared

### Alternative Approaches

**For distributed systems, consider:**
- Redis: Shared cache across instances
- Memcached: High-performance distributed cache
- HTTP caching headers: Let browsers/CDNs cache

**For zentube (single instance):**
- In-memory cache is perfect
- No network overhead
- Simple implementation
- Fast access times

---

### 12. Environment-Specific Configuration

### Implementation
**Files:**
- `internal/config/config.go` - Environment-aware config loading
- `configs/config.development.yaml` - Dev config
- `configs/config.staging.yaml` - Staging config
- `configs/config.production.yaml` - Production config

### Why Environment-Specific Config?

Single config:
```yaml
# config.yaml
app:
  port: 8080
database:
  path: "./zentube.db"
```

**Problems:**
- Production uses local database path
- Same settings for dev and prod
- Hard to test different configurations
- Secrets in version control

### Pattern: Environment-Based Config Loading

```go
// internal/config/config.go
package config

type Environment string

const (
    Development Environment = "development"
    Staging     Environment = "staging"
    Production  Environment = "production"
)

// GetEnvironment reads APP_ENV or defaults to development
func GetEnvironment() Environment {
    env := os.Getenv("APP_ENV")
    switch env {
    case "production", "prod":
        return Production
    case "staging", "stage":
        return Staging
    default:
        return Development
    }
}

// LoadConfig loads environment-specific configuration
// Priority: config.<env>.yaml → config.yaml
func LoadConfig(baseFile string) (*Config, error) {
    env := GetEnvironment()
    
    // Try environment-specific config first
    envFile := fmt.Sprintf("config.%s.yaml", env)
    data, err := os.ReadFile(envFile)
    if err != nil {
        // Fall back to base config
        data, err = os.ReadFile(baseFile)
        if err != nil {
            return nil, err
        }
    }
    
    var config Config
    yaml.Unmarshal(data, &config)
    config.App.Environment = env
    
    return &config, nil
}
```

### Environment-Specific Configs

**Development** (`config.development.yaml`):
```yaml
app:
  name: "zentube"
  port: 8080

youtube:
  max_results: 10  # Lower to save quota

database:
  path: "./zentube_dev.db"  # Local SQLite
```

**Staging** (`config.staging.yaml`):
```yaml
app:
  name: "zentube"
  port: 8080

youtube:
  max_results: 15

database:
  path: "./zentube_staging.db"
```

**Production** (`config.production.yaml`):
```yaml
app:
  name: "zentube"
  port: 8080

youtube:
  max_results: 25  # Higher for better UX

database:
  path: "/var/lib/zentube/zentube.db"  # System directory
```

### Environment Variables

Support `.env.<environment>` files:

```go
func LoadEnv() error {
    env := GetEnvironment()
    
    // Try environment-specific .env first
    envFile := fmt.Sprintf(".env.%s", env)
    if err := godotenv.Load(envFile); err == nil {
        return nil
    }
    
    // Fall back to .env
    return godotenv.Load()
}
```

**Files:**
- `.env` - Default/development
- `.env.staging` - Staging secrets
- `.env.production` - Production secrets (often system env vars)

### Integration in main.go

```go
func run() error {
    // Detect environment
    env := config.GetEnvironment()
    
    // Setup logging based on environment
    logger := middleware.NewLogger(string(env))
    
    // Load environment-specific config
    cfg, err := config.LoadConfig("configs/config.yaml")
    if err != nil {
        return err
    }
    
    logger.Info("loaded configuration",
        slog.String("environment", string(cfg.App.Environment)),
        slog.String("database", cfg.Database.Path),
    )
    
    // Set Gin mode based on environment
    if cfg.IsProduction() {
        gin.SetMode(gin.ReleaseMode)
    } else if cfg.IsDevelopment() {
        gin.SetMode(gin.DebugMode)
    }
    
    // ... rest of setup
}
```

### Config Helper Methods

```go
type Config struct {
    App App
    // ...
}

func (c *Config) IsDevelopment() bool {
    return c.App.Environment == Development
}

func (c *Config) IsProduction() bool {
    return c.App.Environment == Production
}

func (c *Config) IsStaging() bool {
    return c.App.Environment == Staging
}
```

### Running in Different Environments

```bash
# Development (default)
go run ./cmd/zentube

# Staging
APP_ENV=staging go run ./cmd/zentube

# Production
APP_ENV=production ./zentube
```

### Benefits

1. **Environment Isolation**: Different settings per environment
2. **Security**: Production secrets separate from dev
3. **Flexibility**: Easy to test production configs
4. **Clarity**: Explicit environment configuration
5. **Default Safety**: Defaults to development

### Configuration Best Practices

**12-Factor App Compliance:**
1. **Config in environment**: Use env vars for secrets
2. **Strict separation**: Dev, staging, prod configs
3. **Never commit secrets**: `.env` in `.gitignore`
4. **Provide examples**: Include `.env.example`

**Security:**
- Production secrets in system env vars (not files)
- Different API keys per environment
- Restrict file permissions on config files
- Rotate secrets regularly

**Organization:**
```
configs/
├── config.yaml              # Fallback/base config
├── config.development.yaml  # Dev settings
├── config.staging.yaml      # Staging settings
└── config.production.yaml   # Prod settings

.env                    # Development secrets (not committed)
.env.example           # Template for developers
.env.staging           # Staging secrets (not committed)
```

### Production Gotchas

- **Missing config files**: Always provide fallback to base config
- **Hardcoded values**: Use env vars for anything that changes between environments
- **Logging verbosity**: Debug logs in dev, structured JSON in prod
- **Database paths**: Ensure production paths have proper permissions
- **Secret rotation**: Update all environment-specific .env files

---

## Summary of Medium Priority Patterns

The medium priority patterns provide essential production capabilities:

1. **Custom Error Types**: Consistent error handling with proper HTTP status codes
2. **Input Validation**: Security and data quality through centralized validation
3. **API Response Caching**: Performance optimization and quota protection
4. **Environment-Specific Configuration**: Proper separation of dev/staging/prod settings

### Key Takeaways

**Error Handling:**
- Type-safe errors with context
- Automatic HTTP status mapping
- Consistent error responses

**Validation:**
- Centralized validation logic
- Security through sanitization
- Clear error messages

**Caching:**
- In-memory TTL cache for API responses
- Thread-safe with RWMutex
- Automatic expiration and cleanup

**Configuration:**
- Environment-based config loading
- Separate configs per environment
- 12-factor app compliance

**All implemented with pure Go** - no external infrastructure required.

**Next Steps:**
- Add integration tests with real dependencies
- Implement distributed caching (Redis) for multi-instance deployments
- Add configuration hot-reload capability
- Enhance with feature flags for gradual rollouts

