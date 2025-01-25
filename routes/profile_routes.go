package routes

import (
	"github.com/datarohit/gopher-social-backend/controllers"
	"github.com/datarohit/gopher-social-backend/middlewares"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// ProfileRoutes defines routes for profile related operations.
//
// Parameters:
//   - router (*gin.RouterGroup): RouterGroup for profile routes under /profile path.
//   - dbPool (*pgxpool.Pool): Pgx connection pool to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - None
//
// Routes:
//   - PUT /profile/update: Route to update user profile. Requires authentication.
func ProfileRoutes(router *gin.RouterGroup, dbPool *pgxpool.Pool, logger *logrus.Logger) {
	profileStore := stores.NewProfileStore(dbPool)
	profileController := controllers.NewProfileController(profileStore, logger)

	profileRouter := router.Group("/profile")
	profileRouter.Use(middlewares.AuthMiddleware(logger))
	profileRouter.PUT("/update", profileController.UpdateProfile)
}
