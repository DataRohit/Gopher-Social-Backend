package routes

import (
	"github.com/datarohit/gopher-social-backend/controllers"
	"github.com/datarohit/gopher-social-backend/middlewares"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// CommentRoutes defines routes for comment related operations.
//
// Parameters:
//   - router (*gin.RouterGroup): RouterGroup for comment routes under /post/:postID/comment path.
//   - dbPool (*pgxpool.Pool): Pgx connection pool to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - None
//
// Routes:
//   - POST /post/:postID/comment/create: Route to create a comment on a post. Requires authentication.
func CommentRoutes(router *gin.RouterGroup, dbPool *pgxpool.Pool, logger *logrus.Logger) {
	commentStore := stores.NewCommentStore(dbPool)
	postStore := stores.NewPostStore(dbPool)
	authStore := stores.NewAuthStore(dbPool)
	commentController := controllers.NewCommentController(commentStore, postStore, authStore, logger)

	commentRouter := router.Group("/post/:postID/comment")
	commentRouter.Use(middlewares.AuthMiddleware(logger))
	commentRouter.POST("/create", commentController.CreateComment)
}
