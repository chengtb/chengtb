package models

import "time"

type Recipe struct {
	RecipeID   int       `gorm:"primaryKey;autoIncrement;column:recipe_id" json:"recipe_id"`
	DishID     int       `gorm:"column:dish_id;not null" json:"dish_id"`
	Portion    int       `gorm:"column:portion;not null" json:"portion"`
	Name       string    `gorm:"column:name;not null;size:100" json:"name"`
	IsEnabled  bool      `gorm:"column:is_enabled;default:true" json:"is_enabled"`
	MaxPortion int       `gorm:"column:max_portion;default:10" json:"max_portion"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Dish *Dish `gorm:"foreignKey:DishID" json:"dish,omitempty"`
}

func (Recipe) TableName() string { return "recipe" }
