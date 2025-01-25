package controllers

import (
	"net/http"

	"github.com/datarohit/gopher-social-backend/database"
	"github.com/datarohit/gopher-social-backend/models"
	"github.com/gin-gonic/gin"
)

type HealthController struct{}

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
