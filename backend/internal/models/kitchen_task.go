package models

import "time"

type KitchenTaskStatus string

const (
	KitchenTaskStatusPending   KitchenTaskStatus = "pending"
	KitchenTaskStatusAssigned  KitchenTaskStatus = "assigned"
	KitchenTaskStatusCooking   KitchenTaskStatus = "cooking"
	KitchenTaskStatusCompleted KitchenTaskStatus = "completed"
	KitchenTaskStatusCancelled KitchenTaskStatus = "cancelled"
)

type KitchenTaskType string

const (
	KitchenTaskTypeNormal KitchenTaskType = "normal"
	KitchenTaskTypeMerged KitchenTaskType = "merged"
)

type KitchenTask struct {
	Model
	Type                 KitchenTaskType   `json:"type" gorm:"type:varchar(20);default:'normal'"`
	RecipeID             uint              `json:"recipe_id" gorm:"not null;index"`
	BatchNo              string            `json:"batch_no" gorm:"type:varchar(64);uniqueIndex"`
	TargetQuantity       int               `json:"target_quantity" gorm:"not null;default:1"`
	Status               KitchenTaskStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	AssignedChefID       *uint             `json:"assigned_chef_id" gorm:"index"`
	ExpectedCompleteTime *time.Time        `json:"expected_complete_time"`
	Recipe               *Recipe           `json:"recipe,omitempty"`
	Chef                 *Chef             `json:"chef,omitempty" gorm:"foreignKey:AssignedChefID"`
	TaskItems            []TaskItem        `json:"task_items,omitempty" gorm:"foreignKey:KitchenTaskID"`
}
