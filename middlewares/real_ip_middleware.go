package middlewares

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

const RealIPKey = "realIP"

// RealIPMiddleware is a middleware that extracts the real client IP address from the request headers.
// It checks for common headers like X-Forwarded-For and X-Real-IP to determine the real IP.
// If none of these headers are present, it falls back to using the RemoteAddr from the request.
// The real IP address is stored in the context and can be accessed using the RealIPKey.
//
// Returns:
//   - gin.HandlerFunc: A middleware function that extracts and sets the real client IP address in the context.
func RealIPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var realIP string

		if xForwardedFor := c.Request.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
			ips := strings.Split(xForwardedFor, ",")
			realIP = strings.TrimSpace(ips[0])
		}

		if realIP == "" {
			if xRealIP := c.Request.Header.Get("X-Real-IP"); xRealIP != "" {
				realIP = strings.TrimSpace(xRealIP)
			}
		}

		if realIP == "" {
			realIP, _, _ = net.SplitHostPort(c.Request.RemoteAddr)
		}

		c.Set(RealIPKey, realIP)
		c.Next()
	}
}
