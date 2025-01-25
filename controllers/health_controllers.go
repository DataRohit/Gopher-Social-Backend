package controllers

import (
	"net/http"

	"github.com/datarohit/gopher-social-backend/database"
	"github.com/datarohit/gopher-social-backend/models"
	"github.com/gin-gonic/gin"
)

type HealthController struct{}

// NewHealthController creates a new HealthController.
//
// Parameters:
//   - None
//
// Returns:
//   - *HealthController: Pointer to the HealthController.
func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthRouter godoc
// @Summary      Router Health Check
// @Description  Check if the router is working
// @Tags         health
// @Produce      json
// @Success      200 {object} models.RouterHealthyResponse "Successfully connected to router"
// @Router       /health/router [get]
func (hc *HealthController) HealthRouter(c *gin.Context) {
	c.JSON(http.StatusOK, models.RouterHealthyResponse{
		Status: "Router Healthy!",
	})
}

// HealthRedis godoc
// @Summary      Redis Health Check
// @Description  Check if Redis connection is healthy
// @Tags         health
// @Produce      json
// @Success      200 {object} models.RedisHealthyResponse "Successfully connected to Redis"
// @Failure      503 {object} models.RedisUnhealthyResponse "Failed to connect to Redis"
// @Router       /health/redis [get]
func (hc *HealthController) HealthRedis(c *gin.Context) {
	if database.RedisClient != nil && database.RedisClient.Ping(c).Err() == nil {
		c.JSON(http.StatusOK, models.RedisHealthyResponse{
			Status: "Redis Healthy!",
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, models.RedisUnhealthyResponse{
			Status: "Redis Unhealthy!",
		})
	}
}

// HealthPostgres godoc
// @Summary      Postgres Health Check
// @Description  Check if Postgres connection is healthy
// @Tags         health
// @Produce      json
// @Success      200 {object} models.PostgresHealthyResponse "Successfully connected to Postgres"
// @Failure      503 {object} models.PostgresUnhealthyResponse "Failed to connect to Postgres"
// @Router       /health/postgres [get]
func (hc *HealthController) HealthPostgres(c *gin.Context) {
	if database.PostgresDB != nil {
		err := database.PostgresDB.Ping(c)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, models.PostgresUnhealthyResponse{
				Status: "Postgres Unhealthy!",
			})
			return
		}
		c.JSON(http.StatusOK, models.PostgresHealthyResponse{
			Status: "Postgres Healthy!",
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, models.PostgresUnhealthyResponse{
			Status: "Postgres Unhealthy!",
		})
	}
}
