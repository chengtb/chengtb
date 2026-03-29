package models

import "time"

type Category struct {
	CategoryID int       `gorm:"primaryKey;autoIncrement;column:category_id" json:"category_id"`
	ParentID   int       `gorm:"column:parent_id;default:0" json:"parent_id"`
	Name       string    `gorm:"column:name;not null;size:100" json:"name"`
	SortOrder  int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (Category) TableName() string { return "category" }
