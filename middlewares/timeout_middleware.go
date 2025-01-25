package middlewares

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware is a middleware that sets a timeout for each request.
// If a handler takes longer than the specified timeout, the request will be aborted
// and a 408 Request Timeout error will be returned to the client.
//
// Parameters:
//   - timeout time.Duration: The duration after which the request should timeout.
//
// Returns:
//   - gin.HandlerFunc: A middleware function that enforces request timeouts.
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
		case <-ctx.Done():
			c.AbortWithStatusJSON(408, gin.H{
				"error": "Request Timeout!",
			})
			return
		}
	}
}
