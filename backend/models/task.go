package models

import (
	"database/sql"
	"time"

	"gorm.io/datatypes"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCooking   TaskStatus = "cooking"
	TaskStatusDone      TaskStatus = "done"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type CookingTask struct {
	TaskID       int            `gorm:"primaryKey;autoIncrement;column:task_id" json:"task_id"`
	ChefID       int            `gorm:"column:chef_id;not null" json:"chef_id"`
	RecipeID     int            `gorm:"column:recipe_id;not null" json:"recipe_id"`
	DishID       int            `gorm:"column:dish_id;not null" json:"dish_id"`
	TotalPortion int            `gorm:"column:total_portion;not null" json:"total_portion"`
	MergedFrom   datatypes.JSON `gorm:"column:merged_from;type:json" json:"merged_from"`
	TableIDs     datatypes.JSON `gorm:"column:table_ids;type:json" json:"table_ids"`
	Status       TaskStatus     `gorm:"column:status;type:enum('pending','cooking','done','cancelled');default:'pending'" json:"status"`
	Priority     int            `gorm:"column:priority;default:0" json:"priority"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	CompletedAt  sql.NullTime   `gorm:"column:completed_at" json:"completed_at"`

	Chef   *Chef   `json:"chef,omitempty"`
	Recipe *Recipe `json:"recipe,omitempty"`
	Dish   *Dish   `json:"dish,omitempty"`
}

func (CookingTask) TableName() string { return "cooking_task" }

type SystemConfig struct {
	ConfigKey   string `gorm:"primaryKey;column:config_key;size:100" json:"config_key"`
	ConfigValue string `gorm:"column:config_value;not null;size:500" json:"config_value"`
	Description string `gorm:"column:description;size:255" json:"description"`
}

func (SystemConfig) TableName() string { return "system_config" }
