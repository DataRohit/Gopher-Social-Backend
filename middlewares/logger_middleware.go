package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware is a middleware that logs HTTP requests using Logrus.
// It logs the request ID, real IP, path, status code, and duration of the request.
// The log level is determined based on the HTTP status code:
//   - Status codes 5xx and 4xx are logged as errors.
//   - Status codes 3xx are logged as warnings.
//   - All other status codes are logged as info.
//
// Parameters:
//   - logger (*logrus.Logger): A Logrus logger instance.
//
// Returns:
//   - gin.HandlerFunc: A middleware function that logs HTTP requests.
func LoggerMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)

		requestID, _ := c.Get(RequestIDKey)
		realIP, _ := c.Get(RealIPKey)
		statusCode := c.Writer.Status()
		path := c.Request.URL.Path

		logFields := logrus.Fields{
			"request-id": requestID,
			"real-ip":    realIP,
			"path":       path,
			"status":     statusCode,
			"duration":   duration,
		}

		switch {
		case statusCode >= 500:
			logger.WithFields(logFields).Error("HTTP/HTTPS Request - Server Error")
		case statusCode >= 400:
			logger.WithFields(logFields).Warn("HTTP/HTTPS Request - Client Error")
		case statusCode >= 300:
			logger.WithFields(logFields).Warn("HTTP/HTTPS Request - Redirection")
		default:
			logger.WithFields(logFields).Info(" HTTP/HTTPS Request - Success")
		}
	}
}
