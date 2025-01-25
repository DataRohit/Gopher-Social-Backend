package models

// Router Health Models
type RouterHealthyResponse struct {
	Status string `json:"status" example:"Router Healthy!"`
}

// Redis Health Models
type RedisHealthyResponse struct {
	Status string `json:"status" example:"Redis Healthy!"`
}

type RedisUnhealthyResponse struct {
	Status string `json:"status" example:"Redis Unhealthy!"`
}

// Postgres Health Models
type PostgresHealthyResponse struct {
	Status string `json:"status" example:"Postgres Healthy!"`
}

type PostgresUnhealthyResponse struct {
	Status string `json:"status" example:"Postgres Unhealthy!"`
}
