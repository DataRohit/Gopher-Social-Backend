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

type PostgresHealthyResponse struct {
	Status string `json:"status" example:"Postgres Healthy!"`
}

type PostgresUnhealthyResponse struct {
	Status string `json:"status" example:"Postgres Unhealthy!"`
}
