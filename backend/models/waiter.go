package models

import (
	"database/sql"
	"time"

	"gorm.io/datatypes"
)

type WaiterStatus string

const (
	WaiterStatusOnline  WaiterStatus = "online"
	WaiterStatusOffline WaiterStatus = "offline"
)

type Waiter struct {
	WaiterID  int          `gorm:"primaryKey;autoIncrement;column:waiter_id" json:"waiter_id"`
	Name      string       `gorm:"column:name;not null;size:100" json:"name"`
	Phone     string       `gorm:"column:phone;size:20" json:"phone"`
	AreaID    *int         `gorm:"column:area_id" json:"area_id"`
	Status    WaiterStatus `gorm:"column:status;type:enum('online','offline');default:'offline'" json:"status"`
	CreatedAt time.Time    `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Area *DiningArea `json:"area,omitempty"`
}

func (Waiter) TableName() string { return "waiter" }

type DeliveryTaskStatus string

const (
	DeliveryTaskStatusPending    DeliveryTaskStatus = "pending"
	DeliveryTaskStatusDelivering DeliveryTaskStatus = "delivering"
	DeliveryTaskStatusDone       DeliveryTaskStatus = "done"
	DeliveryTaskStatusRejected   DeliveryTaskStatus = "rejected"
)

type DeliveryTask struct {
	TaskID        int                `gorm:"primaryKey;autoIncrement;column:task_id" json:"task_id"`
	CookingTaskID int                `gorm:"column:cooking_task_id;not null" json:"cooking_task_id"`
	WaiterID      *int               `gorm:"column:waiter_id" json:"waiter_id"`
	TableIDs      datatypes.JSON     `gorm:"column:table_ids;type:json" json:"table_ids"`
	Status        DeliveryTaskStatus `gorm:"column:status;type:enum('pending','delivering','done','rejected');default:'pending'" json:"status"`
	PickupTime    sql.NullTime       `gorm:"column:pickup_time" json:"pickup_time"`
	DeliveredTime sql.NullTime       `gorm:"column:delivered_time" json:"delivered_time"`
	RejectReason  string             `gorm:"column:reject_reason;size:255" json:"reject_reason"`
	CreatedAt     time.Time          `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	CookingTask *CookingTask `json:"cooking_task,omitempty"`
	Waiter      *Waiter      `json:"waiter,omitempty"`
}

func (DeliveryTask) TableName() string { return "delivery_task" }
