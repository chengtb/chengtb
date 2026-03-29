package database

import (
	"restaurant-system/models"
)

func Migrate() error {
	return DB.AutoMigrate(
		&models.Category{},
		&models.Table{},
		&models.Dish{},
		&models.Recipe{},
		&models.Chef{},
		&models.ChefRecipe{},
		&models.CartSession{},
		&models.SessionItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.CookingTask{},
		&models.SystemConfig{},
	)
}
