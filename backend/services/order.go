package services

import (
	"errors"
	"time"

	"restaurant-system/database"
	"restaurant-system/models"
)

func SubmitOrder(sessionID int) (*models.Order, error) {
	var session models.CartSession
	if err := database.DB.Preload("Items.Dish").First(&session, sessionID).Error; err != nil {
		return nil, err
	}
	if session.Status != models.CartSessionActive {
		return nil, errors.New("session is not active")
	}
	if len(session.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	// Calculate total
	var total float64
	for _, item := range session.Items {
		if item.Dish != nil {
			total += item.Dish.Price * float64(item.Quantity)
		}
	}

	order := models.Order{
		TableID:     session.TableID,
		SessionID:   &session.SessionID,
		TotalAmount: total,
		Status:      models.OrderStatusPending,
	}

	if err := database.DB.Create(&order).Error; err != nil {
		return nil, err
	}

	// Create order items from session items
	for _, si := range session.Items {
		oi := models.OrderItem{
			OrderID:  order.OrderID,
			DishID:   si.DishID,
			Quantity: si.Quantity,
			Note:     si.Note,
			Status:   models.OrderItemStatusPending,
		}
		if err := database.DB.Create(&oi).Error; err != nil {
			// Roll back order on item creation failure
			database.DB.Delete(&order)
			return nil, err
		}
	}

	// Mark session as submitted
	database.DB.Model(&session).Update("status", models.CartSessionSubmitted)
	// Update table status
	database.DB.Model(&models.Table{}).Where("table_id = ?", session.TableID).
		Update("status", models.TableStatusWaiting)

	database.DB.Preload("Items.Dish").First(&order, order.OrderID)
	return &order, nil
}

func GetOrder(orderID int) (*models.Order, error) {
	var order models.Order
	err := database.DB.Preload("Items.Dish").Preload("Table").First(&order, orderID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func ListOrders(status string, tableID int) ([]models.Order, error) {
	q := database.DB.Preload("Items.Dish").Preload("Table").Order("created_at DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if tableID > 0 {
		q = q.Where("table_id = ?", tableID)
	}
	var orders []models.Order
	err := q.Find(&orders).Error
	return orders, err
}

func UpdateOrderStatus(orderID int, status models.OrderStatus) (*models.Order, error) {
	var order models.Order
	if err := database.DB.First(&order, orderID).Error; err != nil {
		return nil, err
	}
	order.Status = status
	if err := database.DB.Save(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func RecordPayment(orderID int, method string, isVIP bool) (*models.Order, error) {
	var order models.Order
	if err := database.DB.First(&order, orderID).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	order.PaymentMethod = method
	order.IsVIP = isVIP
	order.PaidAt.Time = now
	order.PaidAt.Valid = true
	order.Status = models.OrderStatusCompleted
	if err := database.DB.Save(&order).Error; err != nil {
		return nil, err
	}
	// Free the table
	database.DB.Model(&models.Table{}).Where("table_id = ?", order.TableID).
		Update("status", models.TableStatusIdle)
	return &order, nil
}
