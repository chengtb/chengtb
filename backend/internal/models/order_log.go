package models

type OrderLog struct {
	Model
	OrderID      uint   `json:"order_id" gorm:"not null;index"`
	OperatorID   uint   `json:"operator_id" gorm:"not null"`
	OperatorType string `json:"operator_type" gorm:"type:varchar(20);not null;comment:customer/waiter/chef/system"`
	Action       string `json:"action" gorm:"type:varchar(50);not null"`
	Content      string `json:"content" gorm:"type:text"`
	Order        *Order `json:"order,omitempty"`
}
