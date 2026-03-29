package router

import (
	"restaurant-system/handlers"
	"restaurant-system/middleware"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	// WebSocket
	r.GET("/ws/cart/:sessionId", handlers.CartWebSocket)

	// Customer APIs
	customer := r.Group("/api/customer")
	{
		customer.GET("/table/:tableId", handlers.GetTableInfo)
		customer.POST("/cart/session", handlers.CreateOrGetCartSession)
		customer.GET("/cart/session/:sessionId", handlers.GetCartSession)
		customer.POST("/cart/session/:sessionId/item", handlers.AddItemToCart)
		customer.PUT("/cart/session/:sessionId/item/:itemId", handlers.UpdateCartItem)
		customer.DELETE("/cart/session/:sessionId/item/:itemId", handlers.RemoveCartItem)
		customer.POST("/cart/session/:sessionId/submit", handlers.SubmitOrder)
		customer.GET("/order/:orderId", handlers.GetOrderStatus)
		customer.GET("/dishes", handlers.GetDishes)
		customer.GET("/dishes/category/:categoryId", handlers.GetDishesByCategory)
	}

	// Merchant APIs
	merchant := r.Group("/api/merchant")
	{
		merchant.GET("/tables", handlers.GetAllTables)
		merchant.GET("/tables/:tableId/order", handlers.GetTableOrder)
		merchant.PUT("/tables/:tableId/status", handlers.UpdateTableStatus)

		merchant.GET("/orders", handlers.ListOrders)
		merchant.GET("/orders/:orderId", handlers.GetOrderDetails)
		merchant.PUT("/orders/:orderId/status", handlers.UpdateOrderStatus)
		merchant.POST("/orders/:orderId/payment", handlers.RecordPayment)
		merchant.POST("/orders/:orderId/dispatch", handlers.DispatchOrder)

		merchant.GET("/chefs", handlers.ListChefs)
		merchant.POST("/chefs", handlers.CreateChef)
		merchant.PUT("/chefs/:chefId", handlers.UpdateChef)

		merchant.GET("/recipes", handlers.ListRecipes)
		merchant.POST("/recipes", handlers.CreateRecipe)
		merchant.PUT("/recipes/:recipeId", handlers.UpdateRecipe)

		merchant.GET("/dishes", handlers.MerchantListDishes)
		merchant.POST("/dishes", handlers.CreateDish)
		merchant.PUT("/dishes/:dishId", handlers.UpdateDish)
		merchant.DELETE("/dishes/:dishId", handlers.DeleteDish)

		merchant.GET("/categories", handlers.ListCategories)
		merchant.POST("/categories", handlers.CreateCategory)

		merchant.GET("/tasks", handlers.ListTasks)
		merchant.PUT("/tasks/:taskId/reassign", handlers.ReassignTask)

		merchant.GET("/config", handlers.GetConfig)
		merchant.PUT("/config", handlers.UpdateConfig)
	}

	// Chef APIs
	chef := r.Group("/api/chef")
	{
		chef.POST("/auth/login", handlers.ChefLogin)
		chef.GET("/tasks", handlers.GetChefTasks)
		chef.GET("/tasks/:taskId", handlers.GetChefTaskDetails)
		chef.PUT("/tasks/:taskId/complete", handlers.CompleteTask)
	}

	return r
}
