package middleware

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter manages rate limiters per IP address
type IPRateLimiter struct {
	ips    map[string]*rate.Limiter
	mu     sync.RWMutex
	r      rate.Limit // requests per second
	b      int        // burst size
	logger *slog.Logger
}

// NewIPRateLimiter creates a new IP-based rate limiter
// r: requests per second (e.g., 10 means 10 requests/second)
// b: burst size (e.g., 20 allows bursts up to 20 requests)
func NewIPRateLimiter(r rate.Limit, b int, logger *slog.Logger) *IPRateLimiter {
	return &IPRateLimiter{
		ips:    make(map[string]*rate.Limiter),
		r:      r,
		b:      b,
		logger: logger,
	}
}

// GetLimiter returns the rate limiter for the given IP
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

// RateLimit creates a rate limiting middleware
// Example: RateLimit(10, 20) allows 10 req/sec with bursts up to 20
func RateLimit(r rate.Limit, b int, logger *slog.Logger) gin.HandlerFunc {
	limiter := NewIPRateLimiter(r, b, logger)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiterForIP := limiter.GetLimiter(ip)

		if !limiterForIP.Allow() {
			logger.Warn("rate limit exceeded",
				slog.String("ip", ip),
				slog.String("path", c.Request.URL.Path),
			)

			c.Header("X-RateLimit-Limit", "10")
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "60")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
