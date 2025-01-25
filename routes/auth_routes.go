package routes

import (
	"github.com/datarohit/gopher-social-backend/controllers"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthRoutes function to define authentication routes.
//
// Parameters:
//   - router (*gin.RouterGroup): RouterGroup pointer to define routes under /auth path.
//   - dbPool (*pgxpool.Pool): Pgx connection pool to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - None
//
// Routes:
//   - /auth/register (POST):  Route to register a new user.
func AuthRoutes(router *gin.RouterGroup, dbPool *pgxpool.Pool, logger *logrus.Logger) {
	authStore := stores.NewAuthStore(dbPool)
	authController := controllers.NewAuthController(authStore, logger)

	authRouter := router.Group("/auth")
	authRouter.POST("/register", authController.Register)
}
