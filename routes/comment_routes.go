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
//   - PUT /post/:postID/comment/:commentID/update: Route to update a comment on a post. Requires authentication.
//   - DELETE /post/:postID/comment/:commentID/delete: Route to delete a comment on a post. Requires authentication.
//   - GET /post/:postID/comment/:commentID: Route to get a comment by comment ID and post ID. No authentication required.
//   - GET /post/:postID/comment/user/me: Route to list all comments of logged in user for a post. Requires authentication.
//   - GET /post/:postID/comment/user/:identifier: Route to list all comments of a user for a post. No authentication required.
func CommentRoutes(router *gin.RouterGroup, dbPool *pgxpool.Pool, logger *logrus.Logger) {
	commentStore := stores.NewCommentStore(dbPool)
	postStore := stores.NewPostStore(dbPool)
	authStore := stores.NewAuthStore(dbPool)
	commentController := controllers.NewCommentController(commentStore, postStore, authStore, logger)

	commentRouter := router.Group("/post/:postID/comment")
	commentRouter.Use(middlewares.AuthMiddleware(logger))
	commentRouter.POST("/create", commentController.CreateComment)
	commentRouter.PUT("/:commentID/update", commentController.UpdateComment)
	commentRouter.DELETE("/:commentID/delete", commentController.DeleteComment)
	commentRouter.GET("/:commentID", commentController.GetComment)
	commentRouter.GET("/user/me", middlewares.PaginationMiddleware(), commentController.ListMyComments)
	commentRouter.GET("/user/:identifier", middlewares.PaginationMiddleware(), commentController.ListCommentsByUserIdentifier)
}
