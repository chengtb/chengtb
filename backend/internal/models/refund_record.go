package models

type RefundType string

const (
	RefundTypeNotMade       RefundType = "not_made"
	RefundTypeMadeNotServed RefundType = "made_not_served"
	RefundTypeServed        RefundType = "served"
)

type RefundStatus string

const (
	RefundStatusPending  RefundStatus = "pending"
	RefundStatusApproved RefundStatus = "approved"
	RefundStatusRejected RefundStatus = "rejected"
)

type RefundRecord struct {
	Model
	OrderItemID  uint         `json:"order_item_id" gorm:"not null;index"`
	RefundReason string       `json:"refund_reason" gorm:"type:text;not null"`
	RefundType   RefundType   `json:"refund_type" gorm:"type:varchar(30);not null"`
	Status       RefundStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	ApproverID   *uint        `json:"approver_id"`
	ApprovalCode string       `json:"approval_code,omitempty" gorm:"type:varchar(100)"`
	OrderItem    *OrderItem   `json:"order_item,omitempty"`
	Approver     *Staff       `json:"approver,omitempty" gorm:"foreignKey:ApproverID"`
}
