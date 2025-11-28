package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders adds security-focused HTTP headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking attacks
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS protection (legacy browsers)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content Security Policy - restrict resource loading
		// Adjust based on your needs (e.g., allow YouTube embeds)
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' https://i.ytimg.com; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		// Referrer Policy - control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy - control browser features
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// HSTS adds Strict-Transport-Security header for HTTPS enforcement
// Only use in production with proper HTTPS setup
func HSTS(maxAge int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}
