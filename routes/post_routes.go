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
//   - POST /post: Route to create a new post. Requires authentication.
func PostRoutes(router *gin.RouterGroup, dbPool *pgxpool.Pool, logger *logrus.Logger) {
	authStore := stores.NewAuthStore(dbPool)
	postStore := stores.NewPostStore(dbPool)
	postController := controllers.NewPostController(postStore, authStore, logger)

	postRouter := router.Group("/post")
	postRouter.Use(middlewares.AuthMiddleware(logger))
	postRouter.POST("/create", postController.CreatePost)
}
