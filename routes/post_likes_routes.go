package routes

import (
	"github.com/datarohit/gopher-social-backend/controllers"
	"github.com/datarohit/gopher-social-backend/middlewares"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// PostLikeRoutes defines routes for post like operations.
//
// Parameters:
//   - router (*gin.RouterGroup): RouterGroup for post like routes under /post path.
//   - dbPool (*pgxpool.Pool): Pgx connection pool to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - None
//
// Routes:
//   - POST /post/:postID/like: Route to like a post. Requires authentication.
//   - DELETE /post/:postID/unlike: Route to unlike a post. Requires authentication.
//   - POST /post/:postID/dislike: Route to dislike a post. Requires authentication.
//   - DELETE /post/:postID/undislike: Route to undislike a post. Requires authentication.
//   - GET /post/liked: Route to get all liked posts by logged-in user. Requires authentication.
//   - GET /post/disliked: Route to get all disliked posts by logged-in user. Requires authentication.
func PostLikeRoutes(router *gin.RouterGroup, dbPool *pgxpool.Pool, logger *logrus.Logger) {
	authStore := stores.NewAuthStore(dbPool)
	postStore := stores.NewPostStore(dbPool, authStore)
	postLikesStore := stores.NewPostLikeStore(dbPool)
	postLikesController := controllers.NewPostLikesController(postLikesStore, postStore, authStore, logger)

	postLikeRouter := router.Group("/post")
	postLikeRouter.Use(middlewares.AuthMiddleware(logger))
	postLikeRouter.POST("/:postID/like", postLikesController.LikePost)
	postLikeRouter.DELETE("/:postID/unlike", postLikesController.UnlikePost)
	postLikeRouter.POST("/:postID/dislike", postLikesController.DislikePost)
	postLikeRouter.DELETE("/:postID/undislike", postLikesController.UndislikePost)
	postLikeRouter.GET("/liked", postLikesController.ListLikedPosts)
	postLikeRouter.GET("/disliked", postLikesController.ListDislikedPosts)
}
