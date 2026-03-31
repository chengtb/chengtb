package models

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCooking   OrderStatus = "cooking"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusPaid      OrderStatus = "paid"
)

type Order struct {
	Model
	TableID        uint        `json:"table_id" gorm:"not null;index"`
	WaiterID       *uint       `json:"waiter_id" gorm:"index"`
	OrderSN        string      `json:"order_sn" gorm:"type:varchar(64);not null;uniqueIndex"`
	Status         OrderStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	TotalAmount    float64     `json:"total_amount" gorm:"type:decimal(10,2);default:0"`
	DiscountAmount float64     `json:"discount_amount" gorm:"type:decimal(10,2);default:0"`
	PaidAmount     float64     `json:"paid_amount" gorm:"type:decimal(10,2);default:0"`
	Remark         string      `json:"remark" gorm:"type:text"`
	Table          *Table      `json:"table,omitempty"`
	Staff          *Staff      `json:"waiter,omitempty" gorm:"foreignKey:WaiterID"`
	Items          []OrderItem `json:"items,omitempty" gorm:"foreignKey:OrderID"`
}
