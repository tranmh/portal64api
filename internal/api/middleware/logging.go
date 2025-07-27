package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware returns a logging middleware
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate request processing time
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get status code
		statusCode := c.Writer.Status()

		// Get method
		method := c.Request.Method

		// Build full path
		if raw != "" {
			path = path + "?" + raw
		}

		// Log request
		log.Printf("[%s] %d %s %s %v %s",
			method,
			statusCode,
			clientIP,
			path,
			latency,
			c.Request.UserAgent(),
		)
	}
}

// ErrorHandlingMiddleware handles panics and errors
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(500, gin.H{
					"success": false,
					"error":   "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
