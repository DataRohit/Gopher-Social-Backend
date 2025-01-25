package helpers

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	accessTokenSecret     = GetEnv("JWT_ACCESS_SECRET", "06dcdc54085a52a61eac2c085cea9d9ef05c239594f618d1ca72aee91f315563")
	refreshTokenSecret    = GetEnv("JWT_REFRESH_SECRET", "3b69f710a00d78ed724b6d26953f440d0beca2752762f7b2f546a6a27557137f")
	passwordResetSecret   = GetEnv("JWT_RESET_SECRET", "e07ab84db119a5948e07e78be81ae7e2d6a29c872a1bf301225ccada3ff1c457")
	activationTokenSecret = GetEnv("JWT_ACTIVATION_SECRET", "89be6deda2f0acdd570fde648b271e3f697fa05e51a24ccc8624c8d3bf7c56ab")
	accessTokenExpiry     = 30 * time.Minute
	refreshTokenExpiry    = 6 * time.Hour
	passwordResetExpiry   = 15 * time.Minute
	activationTokenExpiry = 15 * time.Minute
)

// GenerateAccessToken generates a new JWT access token.
//
// Parameters:
//   - userID (uuid.UUID): User ID for whom to generate the token.
//
// Returns:
//   - string: JWT access token.
//   - error: An error if token generation fails.
func GenerateAccessToken(userID uuid.UUID) (string, error) {
	return generateToken(userID, accessTokenSecret, accessTokenExpiry)
}

// GenerateRefreshToken generates a new JWT refresh token.
//
// Parameters:
//   - userID (uuid.UUID): User ID for whom to generate the token.
//
// Returns:
//   - string: JWT refresh token.
//   - error: An error if token generation fails.
func GenerateRefreshToken(userID uuid.UUID) (string, error) {
	return generateToken(userID, refreshTokenSecret, refreshTokenExpiry)
}

// GeneratePasswordResetToken generates a new JWT password reset token.
func GeneratePasswordResetToken(userID uuid.UUID) (string, error) {
	return generateToken(userID, passwordResetSecret, passwordResetExpiry)
}

// GenerateActivationToken generates a new JWT activation token.
func GenerateActivationToken(userID uuid.UUID) (string, error) {
	return generateToken(userID, activationTokenSecret, activationTokenExpiry)
}

// generateToken is a helper function to generate JWT tokens.
//
// Parameters:
//   - userID (uuid.UUID): User ID for whom to generate the token.
//   - secretKey (string): Secret key to sign the token.
//   - expiry (time.Duration): Token expiry duration.
//
// Returns:
//   - string: JWT token.
//   - error: An error if token generation fails.
func generateToken(userID uuid.UUID, secretKey string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// VerifyAccessToken verifies the JWT access token.
//
// Parameters:
//   - tokenString (string): JWT access token string.
//
// Returns:
//   - *jwt.Token: Parsed JWT token if valid.
//   - error: An error if token verification fails.
func VerifyAccessToken(tokenString string) (*jwt.Token, error) {
	return verifyToken(tokenString, accessTokenSecret)
}

// VerifyRefreshToken verifies the JWT refresh token.
//
// Parameters:
//   - tokenString (string): JWT refresh token string.
//
// Returns:
//   - *jwt.Token: Parsed JWT token if valid.
//   - error: An error if token verification fails.
func VerifyRefreshToken(tokenString string) (*jwt.Token, error) {
	return verifyToken(tokenString, refreshTokenSecret)
}

// VerifyPasswordResetToken verifies the JWT password reset token.
func VerifyPasswordResetToken(tokenString string) (*jwt.Token, error) {
	return verifyToken(tokenString, passwordResetSecret)
}

// VerifyActivationToken verifies the JWT activation token.
func VerifyActivationToken(tokenString string) (*jwt.Token, error) {
	return verifyToken(tokenString, activationTokenSecret)
}

// verifyToken is a helper function to verify JWT tokens.
//
// Parameters:
//   - tokenString (string): JWT token string.
//   - secretKey (string): Secret key to verify the token.
//
// Returns:
//   - *jwt.Token: Parsed JWT token if valid.
//   - error: An error if token verification fails.
func verifyToken(tokenString string, secretKey string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return token, nil
}

// ExtractUserIDFromToken extracts user ID from a valid JWT token.
//
// Parameters:
//   - token *jwt.Token: Valid JWT token.
//
// Returns:
//   - uuid.UUID: User ID extracted from the token.
//   - error: An error if user ID extraction fails.
func ExtractUserIDFromToken(token *jwt.Token) (uuid.UUID, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token or claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user_id claim not found or invalid")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse user_id as UUID: %w", err)
	}

	return userID, nil
}
