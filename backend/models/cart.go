package models

import "time"

type CartSessionStatus string

const (
	CartSessionActive    CartSessionStatus = "active"
	CartSessionSubmitted CartSessionStatus = "submitted"
	CartSessionClosed    CartSessionStatus = "closed"
)

type CartSession struct {
	SessionID int               `gorm:"primaryKey;autoIncrement;column:session_id" json:"session_id"`
	TableID   int               `gorm:"column:table_id;not null" json:"table_id"`
	Status    CartSessionStatus `gorm:"column:status;type:enum('active','submitted','closed');default:'active'" json:"status"`
	CreatedAt time.Time         `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Table *Table        `json:"table,omitempty"`
	Items []SessionItem `gorm:"foreignKey:SessionID" json:"items,omitempty"`
}

func (CartSession) TableName() string { return "cart_session" }

type SessionItem struct {
	ItemID    int       `gorm:"primaryKey;autoIncrement;column:item_id" json:"item_id"`
	SessionID int       `gorm:"column:session_id;not null" json:"session_id"`
	DishID    int       `gorm:"column:dish_id;not null" json:"dish_id"`
	Quantity  int       `gorm:"column:quantity;default:1" json:"quantity"`
	Note      string    `gorm:"column:note;size:255" json:"note"`
	AddedBy   string    `gorm:"column:added_by;size:50" json:"added_by"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Dish *Dish `json:"dish,omitempty"`
}

func (SessionItem) TableName() string { return "session_item" }
