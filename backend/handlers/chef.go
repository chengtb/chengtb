package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"restaurant-system/database"
	"restaurant-system/models"

	"github.com/gin-gonic/gin"
)

// ChefLogin POST /api/chef/auth/login
func ChefLogin(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var chef models.Chef
	if err := database.DB.Where("name = ? AND is_active = 1", req.Name).First(&chef).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "chef not found or inactive"})
		return
	}
	c.JSON(http.StatusOK, chef)
}

// GetChefTasks GET /api/chef/tasks  (chef_id via query param)
func GetChefTasks(c *gin.Context) {
	chefID, err := strconv.Atoi(c.Query("chef_id"))
	if err != nil || chefID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chef_id required"})
		return
	}
	var tasks []models.CookingTask
	database.DB.Where("chef_id = ? AND status IN ('pending','cooking')", chefID).
		Preload("Recipe").Preload("Dish").
		Order("priority DESC, created_at ASC").
		Find(&tasks)
	c.JSON(http.StatusOK, tasks)
}

// GetChefTaskDetails GET /api/chef/tasks/:taskId
func GetChefTaskDetails(c *gin.Context) {
	taskID, _ := strconv.Atoi(c.Param("taskId"))
	var task models.CookingTask
	if err := database.DB.Preload("Chef").Preload("Recipe").Preload("Dish").
		First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// CompleteTask PUT /api/chef/tasks/:taskId/complete
func CompleteTask(c *gin.Context) {
	taskID, _ := strconv.Atoi(c.Param("taskId"))
	var task models.CookingTask
	if err := database.DB.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	now := time.Now()
	task.Status = models.TaskStatusDone
	task.CompletedAt.Time = now
	task.CompletedAt.Valid = true
	database.DB.Save(&task)

	// Reduce chef load
	database.DB.Model(&models.Chef{}).Where("chef_id = ?", task.ChefID).
		UpdateColumn("current_load", database.DB.Raw("GREATEST(0, current_load - ?)", task.TotalPortion))

	// Mark the order items from merged_from as done
	var mergedItemIDs []int
	if jsonErr := json.Unmarshal(task.MergedFrom, &mergedItemIDs); jsonErr == nil && len(mergedItemIDs) > 0 {
		database.DB.Model(&models.OrderItem{}).
			Where("item_id IN ? AND status = 'dispatched'", mergedItemIDs).
			Update("status", models.OrderItemStatusDone)
	}

	database.DB.Preload("Chef").Preload("Recipe").Preload("Dish").First(&task, taskID)
	c.JSON(http.StatusOK, task)
}
