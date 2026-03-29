package models

import (
	"database/sql"
	"time"
)

type OrderStatus string
type OrderItemStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCooking   OrderStatus = "cooking"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"

	OrderItemStatusPending    OrderItemStatus = "pending"
	OrderItemStatusDispatched OrderItemStatus = "dispatched"
	OrderItemStatusCooking    OrderItemStatus = "cooking"
	OrderItemStatusDone       OrderItemStatus = "done"
)

type Order struct {
	OrderID       int            `gorm:"primaryKey;autoIncrement;column:order_id" json:"order_id"`
	TableID       int            `gorm:"column:table_id;not null" json:"table_id"`
	SessionID     *int           `gorm:"column:session_id" json:"session_id"`
	TotalAmount   float64        `gorm:"column:total_amount;type:decimal(10,2)" json:"total_amount"`
	Status        OrderStatus    `gorm:"column:status;type:enum('pending','cooking','completed','cancelled');default:'pending'" json:"status"`
	PaidAt        sql.NullTime   `gorm:"column:paid_at" json:"paid_at"`
	PaymentMethod string         `gorm:"column:payment_method;size:50" json:"payment_method"`
	IsVIP         bool           `gorm:"column:is_vip;default:false" json:"is_vip"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Table  *Table      `gorm:"foreignKey:TableID" json:"table,omitempty"`
	Items  []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

func (Order) TableName() string { return "order" }

type OrderItem struct {
	ItemID    int             `gorm:"primaryKey;autoIncrement;column:item_id" json:"item_id"`
	OrderID   int             `gorm:"column:order_id;not null" json:"order_id"`
	DishID    int             `gorm:"column:dish_id;not null" json:"dish_id"`
	Quantity  int             `gorm:"column:quantity;default:1" json:"quantity"`
	Note      string          `gorm:"column:note;size:255" json:"note"`
	Status    OrderItemStatus `gorm:"column:status;type:enum('pending','dispatched','cooking','done');default:'pending'" json:"status"`
	CreatedAt time.Time       `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Dish  *Dish  `gorm:"foreignKey:DishID" json:"dish,omitempty"`
	Order *Order `gorm:"foreignKey:OrderID" json:"-"`
}

func (OrderItem) TableName() string { return "order_item" }
