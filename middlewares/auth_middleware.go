package middlewares

import (
	"errors"
	"net/http"
	"time"

	"github.com/datarohit/gopher-social-backend/database"
	"github.com/datarohit/gopher-social-backend/helpers"
	"github.com/datarohit/gopher-social-backend/models"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware is a middleware function to authenticate user requests using JWT tokens from cookies.
// It checks for access token and refresh token cookies, verifies them, and sets the user in the context.
// It also handles access token refreshing using refresh token if access token is expired.
//
// Parameters:
//   - logger (*logrus.Logger): Logrus logger instance for logging.
//
// Returns:
//   - gin.HandlerFunc: Gin middleware handler function.
func AuthMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessTokenCookie, errAccessToken := c.Cookie("access_token")
		refreshTokenCookie, errRefreshToken := c.Cookie("refresh_token")

		authStore := stores.NewAuthStore(database.PostgresDB)
		var user *models.User

		if errAccessToken == nil {
			accessToken, err := helpers.VerifyAccessToken(accessTokenCookie)
			if err == nil && accessToken.Valid {
				userID, err := helpers.ExtractUserIDFromToken(accessToken)
				if err != nil {
					logger.WithFields(logrus.Fields{"error": err}).Warn("Failed to extract User ID from Access Token")
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "error": "invalid access token"})
					return
				}

				user, err = authStore.GetUserByID(c, userID)
				if err != nil {
					if errors.Is(err, stores.ErrUserNotFound) {
						logger.WithFields(logrus.Fields{"userID": userID}).Warn("User not found from access token's User ID")
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "error": "user not found"})
						return
					}
					logger.WithFields(logrus.Fields{"error": err, "userID": userID}).Error("Failed to get user by ID from access token")
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "error": "internal server error"})
					return
				}
			}
		}

		if user == nil {
			if errRefreshToken != nil {
				logger.Warn("Unauthorized access attempt: No valid access or refresh token found")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "error": "missing auth tokens"})
				return
			}

			refreshToken, err := helpers.VerifyRefreshToken(refreshTokenCookie)
			if err != nil || !refreshToken.Valid {
				logger.WithFields(logrus.Fields{"error": err}).Warn("Invalid refresh token")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "error": "invalid refresh token"})
				return
			}

			userID, err := helpers.ExtractUserIDFromToken(refreshToken)
			if err != nil {
				logger.WithFields(logrus.Fields{"error": err}).Warn("Failed to extract User ID from Refresh Token")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "error": "invalid refresh token"})
				return
			}

			user, err = authStore.GetUserByID(c, userID)
			if err != nil {
				if errors.Is(err, stores.ErrUserNotFound) {
					logger.WithFields(logrus.Fields{"userID": userID}).Warn("User not found from refresh token's User ID")
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "error": "user not found"})
					return
				}
				logger.WithFields(logrus.Fields{"error": err, "userID": userID}).Error("Failed to get user by ID from refresh token")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "error": "internal server error"})
				return
			}

			newAccessToken, err := helpers.GenerateAccessToken(user.ID)
			if err != nil {
				logger.WithFields(logrus.Fields{"error": err, "userID": user.ID}).Error("Failed to generate new access token during refresh")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "error": "internal server error"})
				return
			}

			newRefreshToken, err := helpers.GenerateRefreshToken(user.ID)
			if err != nil {
				logger.WithFields(logrus.Fields{"error": err, "userID": user.ID}).Error("Failed to generate new refresh token during refresh")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "error": "internal server error"})
				return
			}

			c.SetCookie("access_token", newAccessToken, int(time.Minute*30/time.Second), "/", "", true, true)
			c.SetCookie("refresh_token", newRefreshToken, int(time.Hour*6/time.Second), "/", "", true, true)
		}

		if user.Banned {
			logger.WithFields(logrus.Fields{"userID": user.ID}).Warn("Banned user attempted authorized action")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Forbidden", "error": "account banned"})
			return
		}

		if !user.IsActive {
			logger.WithFields(logrus.Fields{"userID": user.ID}).Warn("Inactive user attempted authorized action")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Forbidden", "error": "account not active"})
			return
		}

		if user.TimeoutUntil != nil && user.TimeoutUntil.After(time.Now()) {
			logger.WithFields(logrus.Fields{"userID": user.ID, "timeout_until": user.TimeoutUntil}).Warn("User timeout, attempted authorized action")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Forbidden", "error": "account timeout"})
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
