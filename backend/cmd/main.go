package main

import (
	"log"

	"github.com/chengtb/restaurant-kds/internal/config"
	"github.com/chengtb/restaurant-kds/internal/handlers"
	"github.com/chengtb/restaurant-kds/internal/middleware"
	"github.com/chengtb/restaurant-kds/internal/models"
	"github.com/chengtb/restaurant-kds/internal/services"
	ws "github.com/chengtb/restaurant-kds/internal/websocket"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg := config.Load()

	// Initialize database
	models.InitDB(cfg)

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Initialize services
	orderService := services.NewOrderService(models.DB, hub)
	kitchenService := services.NewKitchenService(models.DB, hub, cfg.Business.MergeWindowMinutes)
	refundService := services.NewRefundService(models.DB, &cfg.Business.RefundPermissions)
	giftService := services.NewGiftService(models.DB)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(models.DB)
	tableHandler := handlers.NewTableHandler(models.DB)
	orderHandler := handlers.NewOrderHandler(orderService)
	kitchenHandler := handlers.NewKitchenHandler(kitchenService)
	refundHandler := handlers.NewRefundHandler(refundService)
	giftHandler := handlers.NewGiftHandler(giftService)

	// Setup router
	r := gin.Default()
	r.Use(middleware.CORS())

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Product & Menu
		v1.GET("/categories", productHandler.ListCategories)
		v1.GET("/products", productHandler.ListProducts)
		v1.GET("/products/:id", productHandler.GetProduct)

		// Tables & Regions
		v1.GET("/regions", tableHandler.ListRegions)
		v1.GET("/tables/:id", tableHandler.GetTable)

		// Orders
		v1.POST("/orders", orderHandler.CreateOrder)
		v1.GET("/orders", orderHandler.ListOrders)
		v1.GET("/orders/:id", orderHandler.GetOrder)
		v1.PUT("/orders/:id/confirm", orderHandler.ConfirmOrder)
		v1.PUT("/orders/:id/items", orderHandler.UpdateOrderItems)
		v1.GET("/tables/:table_id/orders", orderHandler.GetOrdersByTable)

		// Kitchen / KDS
		v1.POST("/orders/:id/split", kitchenHandler.SplitOrder)
		v1.GET("/kitchen/tasks", kitchenHandler.ListTasks)
		v1.PUT("/kitchen/tasks/:id/complete", kitchenHandler.CompleteTask)
		v1.PUT("/kitchen/tasks/:id/reassign", kitchenHandler.ReassignTask)

		// Refunds
		v1.POST("/refunds", refundHandler.CreateRefund)
		v1.GET("/refunds", refundHandler.ListRefunds)

		// Gifts
		v1.POST("/gifts", giftHandler.CreateGift)
		v1.GET("/gifts", giftHandler.ListGifts)
	}

	// WebSocket endpoint
	r.GET("/ws", ws.HandleWebSocket(hub))

	// Serve static files for frontend
	r.Static("/customer", "./frontend/customer-app/dist")
	r.Static("/waiter", "./frontend/waiter-app/dist")
	r.Static("/kds", "./frontend/kds-app/dist")

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
