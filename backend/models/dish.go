package models

import "time"

type Dish struct {
	DishID       int       `gorm:"primaryKey;autoIncrement;column:dish_id" json:"dish_id"`
	CategoryID   int       `gorm:"column:category_id" json:"category_id"`
	Name         string    `gorm:"column:name;not null;size:100" json:"name"`
	Price        float64   `gorm:"column:price;not null;type:decimal(10,2)" json:"price"`
	Image        string    `gorm:"column:image;size:255" json:"image"`
	Description  string    `gorm:"column:description;type:text" json:"description"`
	SpecialFlag  bool      `gorm:"column:special_flag;default:false" json:"special_flag"`
	AllowCombine bool      `gorm:"column:allow_combine;default:true" json:"allow_combine"`
	IsAvailable  bool      `gorm:"column:is_available;default:true" json:"is_available"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Category *Category `gorm:"foreignKey:CategoryID;references:CategoryID" json:"category,omitempty"`
}

func (Dish) TableName() string { return "dish" }
