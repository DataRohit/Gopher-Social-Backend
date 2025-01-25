package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "requestID"

// RequestIDMiddleware is a middleware that sets a unique request ID for each request.
// It also sets the request ID in the response header.
// The request ID is stored in the context and can be accessed using the RequestIDKey.
// The request ID is generated using the UUID v4 algorithm.
// The request ID is also set in the response header with the key "X-Request-ID".
//
// Returns:
//   - gin.HandlerFunc: A middleware function that sets a unique request ID for each request.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}
