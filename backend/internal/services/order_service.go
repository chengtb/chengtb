package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chengtb/restaurant-kds/internal/models"
	ws "github.com/chengtb/restaurant-kds/internal/websocket"
	"gorm.io/gorm"
)

type OrderService struct {
	DB  *gorm.DB
	Hub *ws.Hub
}

func NewOrderService(db *gorm.DB, hub *ws.Hub) *OrderService {
	return &OrderService{DB: db, Hub: hub}
}

// GenerateOrderSN generates a unique order serial number
func (s *OrderService) GenerateOrderSN() string {
	now := time.Now()
	return fmt.Sprintf("ORD%s%04d", now.Format("20060102150405"), now.Nanosecond()%10000)
}

// CreateOrderRequest holds the payload for creating an order
type CreateOrderRequest struct {
	TableID uint              `json:"table_id" binding:"required"`
	Remark  string            `json:"remark"`
	Items   []CreateOrderItem `json:"items" binding:"required,min=1"`
}

// CreateOrderItem holds a single item in the order creation request
type CreateOrderItem struct {
	ProductID  uint            `json:"product_id" binding:"required"`
	SpecDetail json.RawMessage `json:"spec_detail"`
	Quantity   int             `json:"quantity" binding:"required,min=1"`
	Remark     string          `json:"remark"`
}

// CreateOrder creates a new order from a customer submission
func (s *OrderService) CreateOrder(req *CreateOrderRequest) (*models.Order, error) {
	order := &models.Order{
		TableID: req.TableID,
		OrderSN: s.GenerateOrderSN(),
		Status:  models.OrderStatusPending,
		Remark:  req.Remark,
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var totalAmount float64
		var items []models.OrderItem

		for _, item := range req.Items {
			var product models.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				return fmt.Errorf("product %d not found", item.ProductID)
			}
			if product.Status != models.ProductStatusActive {
				return fmt.Errorf("product %s is not available", product.Name)
			}

			unitPrice := product.Price
			if item.SpecDetail != nil {
				var specs []struct {
					SpecID     uint    `json:"spec_id"`
					ExtraPrice float64 `json:"extra_price"`
				}
				if err := json.Unmarshal(item.SpecDetail, &specs); err == nil {
					for _, spec := range specs {
						unitPrice += spec.ExtraPrice
					}
				}
			}

			specDetailStr := ""
			if item.SpecDetail != nil {
				specDetailStr = string(item.SpecDetail)
			}

			orderItem := models.OrderItem{
				ProductID:  item.ProductID,
				SpecDetail: specDetailStr,
				UnitPrice:  unitPrice,
				Quantity:   item.Quantity,
				Status:     models.OrderItemStatusPending,
				Remark:     item.Remark,
			}
			items = append(items, orderItem)
			totalAmount += unitPrice * float64(item.Quantity)
		}

		order.TotalAmount = totalAmount
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		for i := range items {
			items[i].OrderID = order.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		order.Items = items

		tx.Model(&models.Table{}).Where("id = ?", req.TableID).Update("status", models.TableStatusOccupied)

		tx.Create(&models.OrderLog{
			OrderID:      order.ID,
			OperatorID:   0,
			OperatorType: "customer",
			Action:       "create",
			Content:      "Order created by customer",
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.Hub.BroadcastToRole(ws.ClientTypeWaiter, ws.Message{
		Type:    "new_order",
		TableID: order.TableID,
		Data:    order,
	})
	s.Hub.BroadcastToTable(order.TableID, ws.Message{
		Type:    "order_created",
		TableID: order.TableID,
		Data:    order,
	})

	return order, nil
}

// ConfirmOrder confirms an order by a waiter
func (s *OrderService) ConfirmOrder(orderID uint, waiterID uint) (*models.Order, error) {
	var order models.Order
	if err := s.DB.Preload("Items").First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.Status != models.OrderStatusPending {
		return nil, fmt.Errorf("order is not in pending status")
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		order.Status = models.OrderStatusConfirmed
		order.WaiterID = &waiterID
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		tx.Create(&models.OrderLog{
			OrderID:      order.ID,
			OperatorID:   waiterID,
			OperatorType: "waiter",
			Action:       "confirm",
			Content:      "Order confirmed by waiter",
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.Hub.BroadcastToTable(order.TableID, ws.Message{
		Type:    "order_confirmed",
		TableID: order.TableID,
		Data:    order,
	})

	return &order, nil
}

// UpdateOrderItemsRequest holds the payload for modifying order items
type UpdateOrderItemsRequest struct {
	AddItems    []CreateOrderItem `json:"add_items"`
	RemoveItems []uint            `json:"remove_item_ids"`
	UpdateItems []struct {
		ItemID   uint `json:"item_id"`
		Quantity int  `json:"quantity"`
	} `json:"update_items"`
}

// UpdateOrderItems allows waiter to modify order items
func (s *OrderService) UpdateOrderItems(orderID uint, waiterID uint, req *UpdateOrderItemsRequest) (*models.Order, error) {
	var order models.Order
	if err := s.DB.Preload("Items").First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.Status != models.OrderStatusPending && order.Status != models.OrderStatusConfirmed {
		return nil, fmt.Errorf("order cannot be modified in current status")
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if len(req.RemoveItems) > 0 {
			tx.Where("id IN ? AND order_id = ?", req.RemoveItems, orderID).Delete(&models.OrderItem{})
		}

		for _, upd := range req.UpdateItems {
			tx.Model(&models.OrderItem{}).Where("id = ? AND order_id = ?", upd.ItemID, orderID).Update("quantity", upd.Quantity)
		}

		for _, item := range req.AddItems {
			var product models.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				return fmt.Errorf("product %d not found", item.ProductID)
			}
			specDetailStr := ""
			if item.SpecDetail != nil {
				specDetailStr = string(item.SpecDetail)
			}
			tx.Create(&models.OrderItem{
				OrderID:    orderID,
				ProductID:  item.ProductID,
				SpecDetail: specDetailStr,
				UnitPrice:  product.Price,
				Quantity:   item.Quantity,
				Status:     models.OrderItemStatusPending,
				Remark:     item.Remark,
			})
		}

		// Recalculate total
		var items []models.OrderItem
		tx.Where("order_id = ?", orderID).Find(&items)
		var total float64
		for _, item := range items {
			total += item.UnitPrice * float64(item.Quantity)
		}
		tx.Model(&order).Update("total_amount", total)

		tx.Create(&models.OrderLog{
			OrderID:      orderID,
			OperatorID:   waiterID,
			OperatorType: "waiter",
			Action:       "update_items",
			Content:      "Order items updated by waiter",
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.DB.Preload("Items").First(&order, orderID)
	return &order, nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(orderID uint) (*models.Order, error) {
	var order models.Order
	if err := s.DB.Preload("Items.Product").Preload("Table").First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}
	return &order, nil
}

// GetOrdersByTable retrieves orders for a table
func (s *OrderService) GetOrdersByTable(tableID uint) ([]models.Order, error) {
	var orders []models.Order
	if err := s.DB.Preload("Items.Product").Where("table_id = ?", tableID).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// ListOrders lists orders with optional status filter
func (s *OrderService) ListOrders(status string, page, pageSize int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := s.DB.Model(&models.Order{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	err := query.Preload("Items.Product").Preload("Table").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&orders).Error

	return orders, total, err
}
