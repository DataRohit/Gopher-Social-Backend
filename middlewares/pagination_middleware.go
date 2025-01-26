package middlewares

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	PageNumberKey = "page_number"
	PageSize      = 10
)

// PaginationMiddleware extracts and validates page number from query parameters.
// It sets the page number in the gin context if valid, otherwise returns a 400 error.
// The page number must be an integer >= 1.
//
// Parameters:
//   - None
//
// Returns:
//   - gin.HandlerFunc: Gin middleware handler for pagination.
func PaginationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		page, err := strconv.Atoi(pageStr)

		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number, page number must be an integer >= 1"})
			c.Abort()
			return
		}

		c.Set(PageNumberKey, page)
		c.Next()
	}
}
