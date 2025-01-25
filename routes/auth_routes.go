package routes

import (
	"github.com/datarohit/gopher-social-backend/controllers"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// AuthRoutes function to define authentication routes.
//
// Parameters:
//   - router (*gin.RouterGroup): RouterGroup pointer to define routes under /auth path.
//   - dbPool (*pgxpool.Pool): Pgx connection pool to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - None
//
// Routes:
//   - /auth/register (POST):  Route to register a new user.
//   - /auth/login (POST): Route to login user and get JWT tokens.
//   - /auth/logout (POST): Route to logout user and invalidate JWT tokens.
//   - /auth/forgot-password (POST): Route to initiate forgot password flow.
//   - /auth/reset-password (POST): Route to reset password using reset token.
//   - /auth/activate (GET): Route to activate user account using activation token.
//   - /auth/resend-activation-link (POST): Route to resend activation link.
func AuthRoutes(router *gin.RouterGroup, dbPool *pgxpool.Pool, logger *logrus.Logger) {
	authStore := stores.NewAuthStore(dbPool)
	profileStore := stores.NewProfileStore(dbPool)
	authController := controllers.NewAuthController(authStore, profileStore, logger)

	authRouter := router.Group("/auth")
	authRouter.POST("/register", authController.Register)
	authRouter.POST("/login", authController.Login)
	authRouter.POST("/logout", authController.Logout)
	authRouter.POST("/forgot-password", authController.ForgotPassword)
	authRouter.POST("/reset-password", authController.ResetPassword)
	authRouter.GET("/activate", authController.ActivateUser)
	authRouter.POST("/resend-activation-link", authController.ResendActivationLink)
}
