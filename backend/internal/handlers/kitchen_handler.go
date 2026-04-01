package handlers

import (
	"strconv"

	"github.com/chengtb/restaurant-kds/internal/middleware"
	"github.com/chengtb/restaurant-kds/internal/services"
	"github.com/gin-gonic/gin"
)

type KitchenHandler struct {
	KitchenService *services.KitchenService
}

func NewKitchenHandler(ks *services.KitchenService) *KitchenHandler {
	return &KitchenHandler{KitchenService: ks}
}

// SplitOrder splits a confirmed order into kitchen tasks
func (h *KitchenHandler) SplitOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid order ID")
		return
	}

	tasks, err := h.KitchenService.SplitOrderToTasks(uint(id))
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.SuccessWithMessage(c, "Order split into tasks", tasks)
}

// ListTasks lists kitchen tasks
func (h *KitchenHandler) ListTasks(c *gin.Context) {
	status := c.Query("status")
	chefID, _ := strconv.ParseUint(c.Query("chef_id"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	tasks, total, err := h.KitchenService.ListTasks(status, uint(chefID), page, pageSize)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"tasks":     tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CompleteTask marks a task as completed
func (h *KitchenHandler) CompleteTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid task ID")
		return
	}

	task, err := h.KitchenService.CompleteTask(uint(id))
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.SuccessWithMessage(c, "Task completed", task)
}

// ReassignTask reassigns a task to a different chef
func (h *KitchenHandler) ReassignTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid task ID")
		return
	}

	var req struct {
		ChefID uint `json:"chef_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	task, err := h.KitchenService.ReassignTask(uint(id), req.ChefID)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.SuccessWithMessage(c, "Task reassigned", task)
}
