package models

type RouterHealthyResponse struct {
	Status string `json:"status" example:"Router Healthy!"`
}

type RedisHealthyResponse struct {
	Status string `json:"status" example:"Redis Healthy!"`
}

type RedisUnhealthyResponse struct {
	Status string `json:"status" example:"Redis Unhealthy!"`
}
