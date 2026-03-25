// Package api registers Gin routes for the User resource.
package api

import (
	appuser "github.com/chengtb/chengtb/internal/application/user"
	"github.com/chengtb/chengtb/internal/infrastructure/messaging"
	"github.com/chengtb/chengtb/internal/interfaces/dto"
	"github.com/gin-gonic/gin"
	"net/http"
)

// UserHandler handles HTTP requests for the User resource.
// It forwards every request to the User micro-service via NATS request-reply.
type UserHandler struct {
	client *messaging.Client
}

// NewUserHandler constructs the handler.
func NewUserHandler(client *messaging.Client) *UserHandler {
	return &UserHandler{client: client}
}

// RegisterRoutes mounts User routes on the provided router group.
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", h.createUser)
	rg.GET("", h.listUsers)
	rg.GET("/:id", h.getUser)
	rg.PUT("/:id", h.updateUser)
	rg.DELETE("/:id", h.deleteUser)
}

// createUser handles POST /users
func (h *UserHandler) createUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, err)
		return
	}
	result, err := h.client.CreateUser(appuser.CreateUserCommand{
		ID:    req.ID,
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, http.StatusCreated, result)
}

// listUsers handles GET /users
func (h *UserHandler) listUsers(c *gin.Context) {
	users, err := h.client.ListUsers()
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, http.StatusOK, users)
}

// getUser handles GET /users/:id
func (h *UserHandler) getUser(c *gin.Context) {
	user, err := h.client.GetUser(appuser.GetUserQuery{ID: c.Param("id")})
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, http.StatusOK, user)
}

// updateUser handles PUT /users/:id
func (h *UserHandler) updateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, err)
		return
	}
	result, err := h.client.UpdateUser(appuser.UpdateUserCommand{
		ID:    c.Param("id"),
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, http.StatusOK, result)
}

// deleteUser handles DELETE /users/:id
func (h *UserHandler) deleteUser(c *gin.Context) {
	if err := h.client.DeleteUser(appuser.DeleteUserCommand{ID: c.Param("id")}); err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, http.StatusOK, nil)
}
