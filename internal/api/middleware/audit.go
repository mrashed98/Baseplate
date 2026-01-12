package middleware

import (
	"github.com/gin-gonic/gin"
)

const (
	ContextIPAddress = "ip_address"
	ContextUserAgent = "user_agent"
)

// AuditMiddleware extracts and sets audit information in context
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use Gin's ClientIP which respects TrustedProxies configuration.
		// NOTE: IP spoofing prevention requires proper TrustedProxies setup in router.
		// Without explicit SetTrustedProxies() configuration, Gin trusts all proxy headers by default.
		// See: https://pkg.go.dev/github.com/gin-gonic/gin#Engine.SetTrustedProxies
		ipAddress := c.ClientIP()

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
