package handlers

import (
	"net/http"
	"strconv"

	"restaurant-system/database"
	"restaurant-system/models"
	"restaurant-system/services"

	"github.com/gin-gonic/gin"
)

// GetTableInfo GET /api/customer/table/:tableId
func GetTableInfo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("tableId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid table id"})
		return
	}
	var table models.Table
	if err := database.DB.First(&table, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "table not found"})
		return
	}
	c.JSON(http.StatusOK, table)
}

// CreateOrGetCartSession POST /api/customer/cart/session
func CreateOrGetCartSession(c *gin.Context) {
	var req struct {
		TableID int `json:"table_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session, err := services.GetOrCreateCartSession(req.TableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, session)
}

// GetCartSession GET /api/customer/cart/session/:sessionId
func GetCartSession(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("sessionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}
	session, err := services.GetCartSession(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	c.JSON(http.StatusOK, session)
}

// AddItemToCart POST /api/customer/cart/session/:sessionId/item
func AddItemToCart(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("sessionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}
	var req struct {
		DishID   int    `json:"dish_id" binding:"required"`
		Quantity int    `json:"quantity"`
		Note     string `json:"note"`
		AddedBy  string `json:"added_by"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}
	item, err := services.AddItemToCart(sessionID, req.DishID, req.Quantity, req.Note, req.AddedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Broadcast update
	session, _ := services.GetCartSession(sessionID)
	if session != nil {
		services.BroadcastCartUpdate(sessionID, session)
	}
	c.JSON(http.StatusOK, item)
}

// UpdateCartItem PUT /api/customer/cart/session/:sessionId/item/:itemId
func UpdateCartItem(c *gin.Context) {
	sessionID, _ := strconv.Atoi(c.Param("sessionId"))
	itemID, _ := strconv.Atoi(c.Param("itemId"))
	var req struct {
		Quantity int `json:"quantity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := services.UpdateCartItem(sessionID, itemID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	session, _ := services.GetCartSession(sessionID)
	if session != nil {
		services.BroadcastCartUpdate(sessionID, session)
	}
	if item == nil {
		c.JSON(http.StatusOK, gin.H{"message": "item removed"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// RemoveCartItem DELETE /api/customer/cart/session/:sessionId/item/:itemId
func RemoveCartItem(c *gin.Context) {
	sessionID, _ := strconv.Atoi(c.Param("sessionId"))
	itemID, _ := strconv.Atoi(c.Param("itemId"))
	if err := services.RemoveCartItem(sessionID, itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	session, _ := services.GetCartSession(sessionID)
	if session != nil {
		services.BroadcastCartUpdate(sessionID, session)
	}
	c.JSON(http.StatusOK, gin.H{"message": "item removed"})
}

// SubmitOrder POST /api/customer/cart/session/:sessionId/submit
func SubmitOrder(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("sessionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}
	order, err := services.SubmitOrder(sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// GetOrderStatus GET /api/customer/order/:orderId
func GetOrderStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("orderId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	order, err := services.GetOrder(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// GetCategories GET /api/customer/categories
func GetCategories(c *gin.Context) {
	var cats []models.Category
	database.DB.Order("sort_order ASC").Find(&cats)
	c.JSON(http.StatusOK, cats)
}

// GetDishes GET /api/customer/dishes
func GetDishes(c *gin.Context) {
	var dishes []models.Dish
	database.DB.Where("is_available = 1").Preload("Category").Find(&dishes)
	c.JSON(http.StatusOK, dishes)
}

// GetDishesByCategory GET /api/customer/dishes/category/:categoryId
func GetDishesByCategory(c *gin.Context) {
	catID, err := strconv.Atoi(c.Param("categoryId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}
	var dishes []models.Dish
	database.DB.Where("category_id = ? AND is_available = 1", catID).Preload("Category").Find(&dishes)
	c.JSON(http.StatusOK, dishes)
}
