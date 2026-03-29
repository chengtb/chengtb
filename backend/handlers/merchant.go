package handlers

import (
	"net/http"
	"strconv"

	"restaurant-system/database"
	"restaurant-system/models"
	"restaurant-system/services"

	"github.com/gin-gonic/gin"
)

// GetAllTables GET /api/merchant/tables
func GetAllTables(c *gin.Context) {
	var tables []models.Table
	database.DB.Find(&tables)
	c.JSON(http.StatusOK, tables)
}

// CreateTable POST /api/merchant/tables
func CreateTable(c *gin.Context) {
	var table models.Table
	if err := c.ShouldBindJSON(&table); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Create(&table).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, table)
}

// UpdateTable PUT /api/merchant/tables/:tableId
func UpdateTable(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("tableId"))
	var table models.Table
	if err := database.DB.First(&table, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "table not found"})
		return
	}
	if err := c.ShouldBindJSON(&table); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	table.TableID = id
	database.DB.Save(&table)
	c.JSON(http.StatusOK, table)
}

// DeleteTable DELETE /api/merchant/tables/:tableId
func DeleteTable(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("tableId"))
	if err := database.DB.Delete(&models.Table{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetTableOrder GET /api/merchant/tables/:tableId/order
func GetTableOrder(c *gin.Context) {
	tableID, _ := strconv.Atoi(c.Param("tableId"))
	var order models.Order
	err := database.DB.Where("table_id = ? AND status NOT IN ('completed','cancelled')", tableID).
		Preload("Items.Dish").Order("created_at DESC").First(&order).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active order"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// UpdateTableStatus PUT /api/merchant/tables/:tableId/status
func UpdateTableStatus(c *gin.Context) {
	tableID, _ := strconv.Atoi(c.Param("tableId"))
	var req struct {
		Status models.TableStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Model(&models.Table{}).Where("table_id = ?", tableID).
		Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// ListOrders GET /api/merchant/orders
func ListOrders(c *gin.Context) {
	status := c.Query("status")
	tableID, _ := strconv.Atoi(c.Query("table_id"))
	orders, err := services.ListOrders(status, tableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// GetOrderDetails GET /api/merchant/orders/:orderId
func GetOrderDetails(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("orderId"))
	order, err := services.GetOrder(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// UpdateOrderStatus PUT /api/merchant/orders/:orderId/status
func UpdateOrderStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("orderId"))
	var req struct {
		Status models.OrderStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := services.UpdateOrderStatus(id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// RecordPayment POST /api/merchant/orders/:orderId/payment
func RecordPayment(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("orderId"))
	var req struct {
		Method string `json:"method" binding:"required"`
		IsVIP  bool   `json:"is_vip"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := services.RecordPayment(id, req.Method, req.IsVIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// DispatchOrder POST /api/merchant/orders/:orderId/dispatch
func DispatchOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("orderId"))
	results, err := services.ManualDispatch(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// --- Chef management ---

// ListChefs GET /api/merchant/chefs
func ListChefs(c *gin.Context) {
	var chefs []models.Chef
	database.DB.Preload("Recipes").Find(&chefs)
	c.JSON(http.StatusOK, chefs)
}

// CreateChef POST /api/merchant/chefs
func CreateChef(c *gin.Context) {
	var chef models.Chef
	if err := c.ShouldBindJSON(&chef); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Create(&chef).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, chef)
}

// UpdateChef PUT /api/merchant/chefs/:chefId
func UpdateChef(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("chefId"))
	var chef models.Chef
	if err := database.DB.First(&chef, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chef not found"})
		return
	}
	if err := c.ShouldBindJSON(&chef); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	chef.ChefID = id
	database.DB.Save(&chef)
	c.JSON(http.StatusOK, chef)
}

// --- Recipe management ---

// ListRecipes GET /api/merchant/recipes
func ListRecipes(c *gin.Context) {
	var recipes []models.Recipe
	// Joins("Dish") uses a SQL JOIN instead of a separate preload query, so
	// the database enforces recipe.dish_id = dish.dish_id, preventing any
	// in-memory mapping mismatch between the top-level dish_id and dish.dish_id.
	database.DB.Joins("Dish").Find(&recipes)
	c.JSON(http.StatusOK, recipes)
}

// CreateRecipe POST /api/merchant/recipes
func CreateRecipe(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Omit("Dish") prevents GORM's FullSaveAssociations from interfering with
	// the dish_id foreign key when the Dish association pointer is nil.
	if err := database.DB.Omit("Dish").Create(&recipe).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, recipe)
}

// UpdateRecipe PUT /api/merchant/recipes/:recipeId
func UpdateRecipe(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("recipeId"))
	var recipe models.Recipe
	if err := database.DB.First(&recipe, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recipe.RecipeID = id
	// Use Updates with an explicit map instead of Save to bypass GORM's
	// FullSaveAssociations, which can fail to persist dish_id when the Dish
	// association is not preloaded (i.e. Dish pointer is nil).
	if err := database.DB.Model(&recipe).Updates(map[string]interface{}{
		"dish_id":    recipe.DishID,
		"name":       recipe.Name,
		"portion":    recipe.Portion,
		"is_enabled": recipe.IsEnabled,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

// DeleteRecipe DELETE /api/merchant/recipes/:recipeId
func DeleteRecipe(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("recipeId"))
	if err := database.DB.Delete(&models.Recipe{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- Dish management ---

// MerchantListDishes GET /api/merchant/dishes
func MerchantListDishes(c *gin.Context) {
	var dishes []models.Dish
	database.DB.Preload("Category").Find(&dishes)
	c.JSON(http.StatusOK, dishes)
}

// CreateDish POST /api/merchant/dishes
func CreateDish(c *gin.Context) {
	var dish models.Dish
	if err := c.ShouldBindJSON(&dish); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Create(&dish).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dish)
}

// UpdateDish PUT /api/merchant/dishes/:dishId
func UpdateDish(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("dishId"))
	var dish models.Dish
	if err := database.DB.First(&dish, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dish not found"})
		return
	}
	if err := c.ShouldBindJSON(&dish); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dish.DishID = id
	database.DB.Save(&dish)
	c.JSON(http.StatusOK, dish)
}

// DeleteDish DELETE /api/merchant/dishes/:dishId
func DeleteDish(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("dishId"))
	if err := database.DB.Delete(&models.Dish{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- Category management ---

// ListCategories GET /api/merchant/categories
func ListCategories(c *gin.Context) {
	var cats []models.Category
	database.DB.Order("sort_order ASC").Find(&cats)
	c.JSON(http.StatusOK, cats)
}

// CreateCategory POST /api/merchant/categories
func CreateCategory(c *gin.Context) {
	var cat models.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Create(&cat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

// UpdateCategory PUT /api/merchant/categories/:categoryId
func UpdateCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("categoryId"))
	var cat models.Category
	if err := database.DB.First(&cat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat.CategoryID = id
	database.DB.Save(&cat)
	c.JSON(http.StatusOK, cat)
}

// DeleteCategory DELETE /api/merchant/categories/:categoryId
func DeleteCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("categoryId"))
	if err := database.DB.Delete(&models.Category{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- Cooking tasks ---

// ListTasks GET /api/merchant/tasks
func ListTasks(c *gin.Context) {
	var tasks []models.CookingTask
	database.DB.Preload("Chef").Preload("Recipe").Preload("Dish").
		Order("priority DESC, created_at ASC").Find(&tasks)
	c.JSON(http.StatusOK, tasks)
}

// ReassignTask PUT /api/merchant/tasks/:taskId/reassign
func ReassignTask(c *gin.Context) {
	taskID, _ := strconv.Atoi(c.Param("taskId"))
	var req struct {
		ChefID int `json:"chef_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := services.ReassignTask(taskID, req.ChefID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

// --- System config ---

// GetConfig GET /api/merchant/config
func GetConfig(c *gin.Context) {
	var configs []models.SystemConfig
	database.DB.Find(&configs)
	result := map[string]string{}
	for _, cfg := range configs {
		result[cfg.ConfigKey] = cfg.ConfigValue
	}
	c.JSON(http.StatusOK, result)
}

// UpdateConfig PUT /api/merchant/config
func UpdateConfig(c *gin.Context) {
	var updates map[string]string
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for k, v := range updates {
		cfg := models.SystemConfig{ConfigKey: k, ConfigValue: v}
		database.DB.Save(&cfg)
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}
