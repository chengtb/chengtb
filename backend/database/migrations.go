package database

import (
	"restaurant-system/models"
)

func Migrate() error {
	// Extend ENUM columns that already exist – safe no-op on first run
	DB.Exec("ALTER TABLE `order` MODIFY COLUMN `status` ENUM('pending','cooking','dining','completed','cancelled') NOT NULL DEFAULT 'pending'")
	DB.Exec("ALTER TABLE `order_item` MODIFY COLUMN `status` ENUM('pending','dispatched','cooking','done','served') NOT NULL DEFAULT 'pending'")

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
		&models.Waiter{},
		&models.DeliveryTask{},
	)
}
