package routes

import (
	"github.com/datarohit/gopher-social-backend/controllers"
	"github.com/datarohit/gopher-social-backend/middlewares"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// ActionRoutes defines routes for action related operations like timeouts.
//
// Parameters:
//   - router (*gin.RouterGroup): RouterGroup for action routes under /action path.
//   - dbPool (*pgxpool.Pool): Pgx connection pool to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - None
//
// Routes:
//   - POST /action/timeout/:userID: Route to timeout a user. Requires moderator or admin role.
//   - DELETE /action/timeout/:userID: Route to remove timeout from a user. Requires moderator or admin role.
func ActionRoutes(router *gin.RouterGroup, dbPool *pgxpool.Pool, logger *logrus.Logger) {
	authStore := stores.NewAuthStore(dbPool)
	actionStore := stores.NewActionStore(dbPool)
	actionController := controllers.NewActionController(actionStore, authStore, logger)

	actionRouter := router.Group("/action")
	actionRouter.Use(middlewares.AuthMiddleware(logger))
	actionRouter.POST("/timeout/:userID", actionController.TimeoutUser)
	actionRouter.DELETE("/timeout/:userID", actionController.RemoveTimeoutUser)
}
