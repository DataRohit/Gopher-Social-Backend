package routes

import (
	"github.com/datarohit/gopher-social-backend/controllers"
	"github.com/datarohit/gopher-social-backend/middlewares"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// CommentLikeRoutes defines routes for comment like operations.
//
// Parameters:
//   - router (*gin.RouterGroup): RouterGroup for comment like routes under /post/:postID/comment/:commentID path.
//   - dbPool (*pgxpool.Pool): Pgx connection pool to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - None
//
// Routes:
//   - POST /post/:postID/comment/:commentID/like: Route to like a comment. Requires authentication.
//   - DELETE /post/:postID/comment/:commentID/like: Route to unlike a comment. Requires authentication.
//   - POST /post/:postID/comment/:commentID/dislike: Route to dislike a comment. Requires authentication.
//   - DELETE /post/:postID/comment/:commentID/dislike: Route to undislike a comment. Requires authentication.
func CommentLikeRoutes(router *gin.RouterGroup, dbPool *pgxpool.Pool, logger *logrus.Logger) {
	authStore := stores.NewAuthStore(dbPool)
	postStore := stores.NewPostStore(dbPool)
	commentStore := stores.NewCommentStore(dbPool)
	commentLikesStore := stores.NewCommentLikeStore(dbPool)
	commentLikesController := controllers.NewCommentLikesController(commentLikesStore, commentStore, postStore, authStore, logger)

	commentLikeRouter := router.Group("/post/:postID/comment/:commentID")
	commentLikeRouter.Use(middlewares.AuthMiddleware(logger))
	commentLikeRouter.POST("/like", commentLikesController.LikeComment)
	commentLikeRouter.DELETE("/like", commentLikesController.UnlikeComment)
	commentLikeRouter.POST("/dislike", commentLikesController.DislikeComment)
	commentLikeRouter.DELETE("/dislike", commentLikesController.UndislikeComment)
}
