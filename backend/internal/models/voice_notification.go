package models

type VoiceNotification struct {
	Model
	TableID uint         `json:"table_id" gorm:"not null;index"`
	TaskID  uint         `json:"task_id" gorm:"not null;index"`
	Content string       `json:"content" gorm:"type:text;not null"`
	Status  string       `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Table   *Table       `json:"table,omitempty"`
	Task    *KitchenTask `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}
