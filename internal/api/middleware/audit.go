package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextIPAddress = "ip_address"
	ContextUserAgent = "user_agent"
)

// AuditMiddleware extracts and sets audit information in context
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract IP address - check X-Forwarded-For first (for proxies)
		ipAddress := c.GetHeader("X-Forwarded-For")
		if ipAddress == "" {
			ipAddress = c.GetHeader("X-Real-IP")
		}
		if ipAddress == "" {
			ipAddress = c.ClientIP()
		}
		// Handle comma-separated IPs (take the first one)
		if idx := strings.Index(ipAddress, ","); idx != -1 {
			ipAddress = strings.TrimSpace(ipAddress[:idx])
		}

		// Extract user agent
		userAgent := c.GetHeader("User-Agent")

		// Set in context
		c.Set(ContextIPAddress, ipAddress)
		c.Set(ContextUserAgent, userAgent)

		c.Next()
	}
}

// GetIPAddress retrieves IP address from context
func GetIPAddress(c *gin.Context) string {
	val, exists := c.Get(ContextIPAddress)
	if !exists {
		return ""
	}
	if ip, ok := val.(string); ok {
		return ip
	}
	return ""
}

// GetUserAgent retrieves user agent from context
func GetUserAgent(c *gin.Context) string {
	val, exists := c.Get(ContextUserAgent)
	if !exists {
		return ""
	}
	if ua, ok := val.(string); ok {
		return ua
	}
	return ""
}
