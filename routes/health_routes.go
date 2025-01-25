package routes

import (
	"github.com/datarohit/gopher-social-backend/controllers"
	"github.com/gin-gonic/gin"
)

// HealthRoutes function to define health routes.
//
// Parameters:
//   - router (*gin.RouterGroup): RouterGroup pointer to define routes.
//
// Returns:
//   - None
//
// Routes:
//   - /health/router (GET): Health check for router.
//   - /health/redis (GET): Health check for redis.
//   - /health/postgres (GET): Health check for postgres.
func HealthRoutes(router *gin.RouterGroup) {
	healthController := controllers.NewHealthController()

	routerHealth := router.Group("/health")
	routerHealth.GET("/router", healthController.HealthRouter)
	routerHealth.GET("/redis", healthController.HealthRedis)
	routerHealth.GET("/postgres", healthController.HealthPostgres)
}
