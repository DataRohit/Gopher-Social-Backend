package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RecovererMiddleware is a middleware that recovers from any panics and logs the error.
// It prevents the server from crashing and returns a 500 Internal Server Error to the client.
//
// Parameters:
//   - logger (*logrus.Logger): A Logrus logger instance.
//
// Returns:
//   - gin.HandlerFunc: A middleware function that recovers from panics and logs errors.
func RecovererMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			requestID, _ := c.Get(RequestIDKey)

			if err := recover(); err != nil {
				logger.WithFields(logrus.Fields{
					"request-id": requestID,
					"error":      err,
				}).Error("Panic Recovered!")

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error!",
				})
			}
		}()

		c.Next()
	}
}
