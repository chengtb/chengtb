package models

import "time"

type TableStatus string

const (
	TableStatusIdle     TableStatus = "idle"
	TableStatusOrdering TableStatus = "ordering"
	TableStatusWaiting  TableStatus = "waiting"
	TableStatusDining   TableStatus = "dining"
	TableStatusCheckout TableStatus = "checkout"
)

type Table struct {
	TableID   int         `gorm:"primaryKey;autoIncrement;column:table_id" json:"table_id"`
	TableNo   string      `gorm:"column:table_no;not null;uniqueIndex;size:50" json:"table_no"`
	Status    TableStatus `gorm:"column:status;type:enum('idle','ordering','waiting','dining','checkout');default:'idle'" json:"status"`
	QRCode    string      `gorm:"column:qr_code;size:255" json:"qr_code"`
	Capacity  int         `gorm:"column:capacity;default:4" json:"capacity"`
	AreaID    *int        `gorm:"column:area_id" json:"area_id"`
	CreatedAt time.Time   `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Area *DiningArea `json:"area,omitempty"`
}

func (Table) TableName() string { return "table" }
