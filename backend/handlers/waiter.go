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

// WaiterLogin POST /api/waiter/auth/login
func WaiterLogin(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var waiter models.Waiter
	if err := database.DB.Where("name = ?", req.Name).First(&waiter).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "waiter not found"})
		return
	}
	c.JSON(http.StatusOK, waiter)
}

// GetWaiterProfile GET /api/waiter/waiters/:waiterId
func GetWaiterProfile(c *gin.Context) {
	waiterID, err := strconv.Atoi(c.Param("waiterId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid waiter id"})
		return
	}
	var waiter models.Waiter
	if err := database.DB.First(&waiter, waiterID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "waiter not found"})
		return
	}
	c.JSON(http.StatusOK, waiter)
}

// GetDeliveryTasks GET /api/waiter/tasks
// Returns tasks relevant to this waiter: all pending + this waiter's delivering tasks
func GetDeliveryTasks(c *gin.Context) {
	raw := c.GetHeader("X-Waiter-ID")
	if raw == "" {
		raw = c.Query("waiter_id")
	}
	waiterID, err := strconv.Atoi(raw)
	if err != nil || waiterID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "waiter_id required"})
		return
	}

	// Get waiter info for area filtering
	var waiter models.Waiter
	if err := database.DB.First(&waiter, waiterID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "waiter not found"})
		return
	}

	var tasks []models.DeliveryTask
	query := database.DB.Preload("CookingTask.Dish").Preload("CookingTask.Recipe").Preload("Waiter")

	if waiter.AreaAssigned != "" {
		// For pending tasks, filter by area if waiter has an area assigned
		// For delivering tasks, show only this waiter's tasks
		query = query.Where(
			"(status = 'pending') OR (status = 'delivering' AND waiter_id = ?)",
			waiterID,
		)
	} else {
		query = query.Where(
			"status = 'pending' OR (status = 'delivering' AND waiter_id = ?)",
			waiterID,
		)
	}

	query.Order("created_at ASC").Find(&tasks)
	c.JSON(http.StatusOK, tasks)
}

// GetDeliveryTaskDetails GET /api/waiter/tasks/:taskId
func GetDeliveryTaskDetails(c *gin.Context) {
	taskID, _ := strconv.Atoi(c.Param("taskId"))
	var task models.DeliveryTask
	if err := database.DB.Preload("CookingTask.Dish").Preload("CookingTask.Recipe").Preload("Waiter").
		First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// PickupDeliveryTask PUT /api/waiter/tasks/:taskId/pickup
// Optimistic lock: only succeeds if status is still 'pending'
func PickupDeliveryTask(c *gin.Context) {
	taskID, _ := strconv.Atoi(c.Param("taskId"))

	raw := c.GetHeader("X-Waiter-ID")
	if raw == "" {
		raw = c.Query("waiter_id")
	}
	waiterID, err := strconv.Atoi(raw)
	if err != nil || waiterID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "waiter_id required"})
		return
	}

	now := time.Now()
	result := database.DB.Model(&models.DeliveryTask{}).
		Where("task_id = ? AND status = ?", taskID, models.DeliveryTaskStatusPending).
		Updates(map[string]interface{}{
			"status":      models.DeliveryTaskStatusDelivering,
			"waiter_id":   waiterID,
			"pickup_time": now,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "task already picked up or not found"})
		return
	}

	var task models.DeliveryTask
	database.DB.Preload("CookingTask.Dish").Preload("CookingTask.Recipe").Preload("Waiter").
		First(&task, taskID)
	c.JSON(http.StatusOK, task)
}

// ConfirmDelivery PUT /api/waiter/tasks/:taskId/deliver
// Waiter confirms the dishes are delivered to the table (no scan QR needed)
func ConfirmDelivery(c *gin.Context) {
	taskID, _ := strconv.Atoi(c.Param("taskId"))

	var task models.DeliveryTask
	if err := database.DB.Preload("CookingTask").First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	if task.Status != models.DeliveryTaskStatusDelivering {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task is not in delivering state"})
		return
	}

	now := time.Now()
	task.Status = models.DeliveryTaskStatusDone
	task.DeliveredTime.Time = now
	task.DeliveredTime.Valid = true
	database.DB.Save(&task)

	// Mark order items from this cooking task as 'served'
	if task.CookingTask != nil {
		var mergedItemIDs []int
		if jsonErr := json.Unmarshal(task.CookingTask.MergedFrom, &mergedItemIDs); jsonErr == nil && len(mergedItemIDs) > 0 {
			database.DB.Model(&models.OrderItem{}).
				Where("item_id IN ? AND status = 'done'", mergedItemIDs).
				Update("status", models.OrderItemStatusServed)

			// Check if all items in each affected order are served → order becomes 'dining'
			var orderIDs []int
			database.DB.Model(&models.OrderItem{}).
				Select("DISTINCT order_id").
				Where("item_id IN ?", mergedItemIDs).
				Pluck("order_id", &orderIDs)

			for _, orderID := range orderIDs {
				var total, served int64
				database.DB.Model(&models.OrderItem{}).
					Where("order_id = ?", orderID).Count(&total)
				database.DB.Model(&models.OrderItem{}).
					Where("order_id = ? AND status = 'served'", orderID).Count(&served)
				if total > 0 && total == served {
					database.DB.Model(&models.Order{}).
						Where("order_id = ? AND status = 'cooking'", orderID).
						Update("status", models.OrderStatusDining)
					// Notify customer via WebSocket
					notifyCustomerDining(orderID)
				}
			}
		}
	}

	database.DB.Preload("CookingTask.Dish").Preload("CookingTask.Recipe").Preload("Waiter").
		First(&task, taskID)
	c.JSON(http.StatusOK, task)
}

// RejectDeliveryTask PUT /api/waiter/tasks/:taskId/reject
// Waiter returns dishes to kitchen with a reason
func RejectDeliveryTask(c *gin.Context) {
	taskID, _ := strconv.Atoi(c.Param("taskId"))
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var task models.DeliveryTask
	if err := database.DB.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	if task.Status == models.DeliveryTaskStatusDone || task.Status == models.DeliveryTaskStatusRejected {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task already completed or rejected"})
		return
	}

	task.Status = models.DeliveryTaskStatusRejected
	task.RejectReason = req.Reason
	task.WaiterID = nil
	database.DB.Save(&task)

	database.DB.Preload("CookingTask.Dish").Preload("CookingTask.Recipe").
		First(&task, taskID)
	c.JSON(http.StatusOK, task)
}

// notifyCustomerDining broadcasts a dining notification to the customer WebSocket session
func notifyCustomerDining(orderID int) {
	var order models.Order
	if err := database.DB.First(&order, orderID).Error; err != nil {
		return
	}
	if order.SessionID == nil {
		return
	}
	msg, _ := json.Marshal(map[string]interface{}{
		"type":     "dining",
		"order_id": orderID,
		"message":  "您的菜品已全部上桌，请享用！",
	})
	services.Hub.BroadcastToSession(*order.SessionID, msg)
}
