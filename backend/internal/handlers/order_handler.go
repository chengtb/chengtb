package handlers

import (
	"strconv"

	"github.com/chengtb/restaurant-kds/internal/middleware"
	"github.com/chengtb/restaurant-kds/internal/services"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	OrderService *services.OrderService
}

func NewOrderHandler(os *services.OrderService) *OrderHandler {
	return &OrderHandler{OrderService: os}
}

// CreateOrder handles customer order submission
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req services.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	order, err := h.OrderService.CreateOrder(&req)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.SuccessWithMessage(c, "Order created successfully", order)
}

// GetOrder returns order details
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid order ID")
		return
	}

	order, err := h.OrderService.GetOrder(uint(id))
	if err != nil {
		middleware.NotFound(c, err.Error())
		return
	}

	middleware.Success(c, order)
}

// GetOrdersByTable returns orders for a specific table
func (h *OrderHandler) GetOrdersByTable(c *gin.Context) {
	tableID, err := strconv.ParseUint(c.Param("table_id"), 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid table ID")
		return
	}

	orders, err := h.OrderService.GetOrdersByTable(uint(tableID))
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.Success(c, orders)
}

// ListOrders lists orders with optional status filter
func (h *OrderHandler) ListOrders(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	orders, total, err := h.OrderService.ListOrders(status, page, pageSize)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"orders":    orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ConfirmOrder handles waiter order confirmation
func (h *OrderHandler) ConfirmOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid order ID")
		return
	}

	var req struct {
		WaiterID uint `json:"waiter_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	order, err := h.OrderService.ConfirmOrder(uint(id), req.WaiterID)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.SuccessWithMessage(c, "Order confirmed", order)
}

// UpdateOrderItems handles waiter modifications
func (h *OrderHandler) UpdateOrderItems(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid order ID")
		return
	}

	var req services.UpdateOrderItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	waiterID, _ := strconv.ParseUint(c.GetHeader("X-Waiter-ID"), 10, 32)

	order, err := h.OrderService.UpdateOrderItems(uint(id), uint(waiterID), &req)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.Success(c, order)
}
