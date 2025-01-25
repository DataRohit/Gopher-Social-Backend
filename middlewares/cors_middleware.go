package middlewares

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware is a middleware that handles Cross-Origin Resource Sharing (CORS).
// It allows requests from different origins to access the API.
// For production, configure this middleware with specific origins and methods.
//
// Returns:
//   - gin.HandlerFunc: A middleware function that handles CORS.
func CORSMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "Authorization"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	return cors.New(config)
}
