package models

type OrderItemStatus string

const (
	OrderItemStatusPending  OrderItemStatus = "pending"
	OrderItemStatusCooking  OrderItemStatus = "cooking"
	OrderItemStatusCooked   OrderItemStatus = "cooked"
	OrderItemStatusServed   OrderItemStatus = "served"
	OrderItemStatusRefunded OrderItemStatus = "refunded"
)

type OrderItem struct {
	Model
	OrderID    uint            `json:"order_id" gorm:"not null;index"`
	ProductID  uint            `json:"product_id" gorm:"not null;index"`
	SpecDetail string          `json:"spec_detail" gorm:"type:json;comment:selected specs JSON"`
	UnitPrice  float64         `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	Quantity   int             `json:"quantity" gorm:"not null;default:1"`
	Status     OrderItemStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Remark     string          `json:"remark" gorm:"type:text"`
	Order      *Order          `json:"order,omitempty"`
	Product    *Product        `json:"product,omitempty"`
}
