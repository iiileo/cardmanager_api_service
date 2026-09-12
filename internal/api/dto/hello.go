package dto

type HelloRequest struct {
	Name string `json:"name" form:"name" binding:"required"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

type HealthResponse struct {
	Status string `json:"status"`
}
