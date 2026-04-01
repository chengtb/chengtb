package models

import (
	"fmt"
	"log"
	"time"

	"github.com/chengtb/restaurant-kds/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Model struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

var DB *gorm.DB

func InitDB(cfg *config.Config) {
	var err error
	DB, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetimeMin) * time.Minute)

	err = DB.AutoMigrate(
		&Region{}, &Table{}, &Category{}, &Product{}, &ProductSpec{},
		&Recipe{}, &Chef{}, &Order{}, &OrderItem{}, &OrderLog{},
		&KitchenTask{}, &TaskItem{}, &VoiceNotification{},
		&RefundRecord{}, &GiftRecord{},
		&Coupon{}, &UserCoupon{}, &Promotion{},
		&Staff{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}
	fmt.Println("Database migrated successfully")
}
