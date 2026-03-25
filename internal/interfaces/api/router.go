package api

import (
	"github.com/chengtb/chengtb/internal/infrastructure/messaging"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts all application routes on the given Gin engine.
func RegisterRoutes(r *gin.Engine, natsClient *messaging.Client) {
	v1 := r.Group("/api/v1")
	userHandler := NewUserHandler(natsClient)
	userHandler.RegisterRoutes(v1.Group("/users"))
}
