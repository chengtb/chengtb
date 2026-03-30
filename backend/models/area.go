package models

import "time"

// DiningArea represents a named seating zone shared by tables and waiters.
type DiningArea struct {
	AreaID      int       `gorm:"primaryKey;autoIncrement;column:area_id" json:"area_id"`
	Name        string    `gorm:"column:name;not null;uniqueIndex;size:100" json:"name"`
	Description string    `gorm:"column:description;size:255" json:"description"`
	SortOrder   int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (DiningArea) TableName() string { return "dining_area" }
