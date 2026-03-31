package models

type GiftTriggerScene string

const (
	GiftSceneCompensation GiftTriggerScene = "compensation"
	GiftSceneCRM          GiftTriggerScene = "crm"
	GiftSceneMarketing    GiftTriggerScene = "marketing"
)

type GiftApprovalStatus string

const (
	GiftApprovalPending  GiftApprovalStatus = "pending"
	GiftApprovalApproved GiftApprovalStatus = "approved"
	GiftApprovalRejected GiftApprovalStatus = "rejected"
	GiftApprovalAuto     GiftApprovalStatus = "auto"
)

type GiftRecord struct {
	Model
	OrderID        uint               `json:"order_id" gorm:"not null;index"`
	ProductID      uint               `json:"product_id" gorm:"not null;index"`
	Quantity       int                `json:"quantity" gorm:"not null;default:1"`
	TriggerScene   GiftTriggerScene   `json:"trigger_scene" gorm:"type:varchar(30);not null"`
	ApprovalStatus GiftApprovalStatus `json:"approval_status" gorm:"type:varchar(20);default:'pending'"`
	OperatorID     *uint              `json:"operator_id"`
	Remark         string             `json:"remark" gorm:"type:text"`
	Order          *Order             `json:"order,omitempty"`
	Product        *Product           `json:"product,omitempty"`
	Operator       *Staff             `json:"operator,omitempty" gorm:"foreignKey:OperatorID"`
}
