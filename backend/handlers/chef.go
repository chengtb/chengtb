package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"restaurant-system/database"
	"restaurant-system/models"
	"restaurant-system/services"

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

// GetChefProfile GET /api/chef/chefs/:chefId
func GetChefProfile(c *gin.Context) {
	chefID, err := strconv.Atoi(c.Param("chefId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chef id"})
		return
	}
	var chef models.Chef
	if err := database.DB.First(&chef, chefID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chef not found"})
		return
	}
	c.JSON(http.StatusOK, chef)
}

// GetChefTasks GET /api/chef/tasks  (chef_id via X-Chef-ID header or chef_id query param)
func GetChefTasks(c *gin.Context) {
	raw := c.GetHeader("X-Chef-ID")
	if raw == "" {
		raw = c.Query("chef_id")
	}
	chefID, err := strconv.Atoi(raw)
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

	// Create a delivery task for the waiter
	deliveryTask := models.DeliveryTask{
		CookingTaskID: task.TaskID,
		TableIDs:      task.TableIDs,
		Status:        models.DeliveryTaskStatusPending,
	}
	if err := database.DB.Create(&deliveryTask).Error; err == nil {
		// Notify connected waiters about the new task
		database.DB.Preload("CookingTask.Dish").Preload("CookingTask.Recipe").
			First(&deliveryTask, deliveryTask.TaskID)
		notifyWaiters(task, deliveryTask)
	}

	database.DB.Preload("Chef").Preload("Recipe").Preload("Dish").First(&task, taskID)
	c.JSON(http.StatusOK, task)
}

// notifyWaiters broadcasts the new delivery task to connected waiters
func notifyWaiters(cookingTask models.CookingTask, deliveryTask models.DeliveryTask) {
	msg, err := json.Marshal(map[string]interface{}{
		"type":          "new_delivery_task",
		"delivery_task": deliveryTask,
	})
	if err != nil {
		return
	}
	// Determine the area based on the first table in table_ids
	var tableIDs []int
	area := ""
	if jsonErr := json.Unmarshal(cookingTask.TableIDs, &tableIDs); jsonErr == nil && len(tableIDs) > 0 {
		var table models.Table
		if err := database.DB.First(&table, tableIDs[0]).Error; err == nil {
			area = table.Area
		}
	}
	services.WaiterHub.BroadcastNewTask(area, msg)
}
