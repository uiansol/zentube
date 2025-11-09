package middleware

import "github.com/gin-gonic/gin"

// HTMX middleware to detect HTMX requests
func HTMX() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set flag if it's an HTMX request
		if c.GetHeader("HX-Request") == "true" {
			c.Set("is_htmx", true)
		}
		c.Next()
	}
}

// IsHTMXRequest checks if the request is from HTMX
func IsHTMXRequest(c *gin.Context) bool {
	val, exists := c.Get("is_htmx")
	if !exists {
		return false
	}
	isHTMX, ok := val.(bool)
	return ok && isHTMX
}
