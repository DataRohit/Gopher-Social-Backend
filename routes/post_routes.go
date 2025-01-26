package routes

import (
	"github.com/datarohit/gopher-social-backend/controllers"
	"github.com/datarohit/gopher-social-backend/middlewares"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// PostRoutes defines routes for post related operations.
//
// Parameters:
//   - router (*gin.RouterGroup): RouterGroup for post routes under /posts path.
//   - dbPool (*pgxpool.Pool): Pgx connection pool to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - None
//
// Routes:
//   - POST /post/create: Route to create a new post. Requires authentication.
//   - PUT /post/:postID: Route to update an existing post. Requires authentication and author role.
//   - DELETE /post/:postID: Route to delete an existing post. Requires authentication and author role.
//   - GET /post/:postID: Route to get a post by ID. Requires authentication.
func PostRoutes(router *gin.RouterGroup, dbPool *pgxpool.Pool, logger *logrus.Logger) {
	authStore := stores.NewAuthStore(dbPool)
	postStore := stores.NewPostStore(dbPool, authStore)
	postController := controllers.NewPostController(postStore, authStore, logger)

	postRouter := router.Group("/post")
	postRouter.Use(middlewares.AuthMiddleware(logger))
	postRouter.POST("/create", postController.CreatePost)
	postRouter.PUT("/:postID", postController.UpdatePost)
	postRouter.DELETE("/:postID", postController.DeletePost)
	postRouter.GET("/:postID", postController.GetPost)
}
