package models

import (
	"fmt"
	"log"
	"time"

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

func InitDB(dsn string) {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

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
