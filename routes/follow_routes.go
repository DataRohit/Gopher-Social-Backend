package routes

import (
	"github.com/datarohit/gopher-social-backend/controllers"
	"github.com/datarohit/gopher-social-backend/middlewares"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// FollowRoutes defines routes for follow related operations.
//
// Parameters:
//   - router (*gin.RouterGroup): RouterGroup for follow routes under /user path.
//   - dbPool (*pgxpool.Pool): Pgx connection pool to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - None
//
// Routes:
//   - POST /user/follow/:identifier: Route to follow a user. Requires authentication.
//   - DELETE /user/unfollow/:identifier: Route to unfollow a user. Requires authentication.
//   - GET /user/followers: Route to get followers of logged in user. Requires authentication.
//   - GET /user/following: Route to get users being followed by logged in user. Requires authentication.
//   - GET /user/:identifier/followers: Route to get followers of a user by identifier. Requires authentication.
//   - GET /user/:identifier/following: Route to get users being followed by user by identifier. Requires authentication.
func FollowRoutes(router *gin.RouterGroup, dbPool *pgxpool.Pool, logger *logrus.Logger) {
	authStore := stores.NewAuthStore(dbPool)
	profileStore := stores.NewProfileStore(dbPool)
	followStore := stores.NewFollowStore(dbPool)
	followController := controllers.NewFollowController(authStore, profileStore, followStore, logger)

	followRouter := router.Group("/user")
	followRouter.Use(middlewares.AuthMiddleware(logger))
	followRouter.POST("/follow/:identifier", followController.FollowUser)
	followRouter.DELETE("/unfollow/:identifier", followController.UnfollowUser)
	followRouter.GET("/followers", followController.GetFollowers)
	followRouter.GET("/following", followController.GetFollowing)
	followRouter.GET("/:identifier/followers", followController.GetUserFollowers)
	followRouter.GET("/:identifier/following", followController.GetUserFollowing)
}
