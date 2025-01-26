package routes

import (
	"github.com/datarohit/gopher-social-backend/controllers"
	"github.com/datarohit/gopher-social-backend/middlewares"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// FeedRoutes defines routes for feed related operations.
//
// Parameters:
//   - router (*gin.RouterGroup): RouterGroup for feed routes under /feed path.
//   - dbPool (*pgxpool.Pool): Pgx connection pool to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - None
//
// Routes:
//   - GET /feed: Route to get latest posts for feed. No authentication required.
//   - GET /feed/:postID: Route to get a specific post with comments for feed. No authentication required.
func FeedRoutes(router *gin.RouterGroup, dbPool *pgxpool.Pool, logger *logrus.Logger) {
	feedStore := stores.NewFeedStore(dbPool)
	feedController := controllers.NewFeedController(feedStore, logger)

	feedRouter := router.Group("/")
	feedRouter.GET("/feed", middlewares.PaginationMiddleware(), feedController.ListFeed)
	feedRouter.GET("/feed/:postID", middlewares.PaginationMiddleware(), feedController.GetFeedPost)
}
