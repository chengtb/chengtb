package models

type TaskItem struct {
	Model
	KitchenTaskID uint         `json:"kitchen_task_id" gorm:"not null;index"`
	OrderItemID   uint         `json:"order_item_id" gorm:"not null;index"`
	TableID       uint         `json:"table_id" gorm:"not null;index"`
	Status        string       `json:"status" gorm:"type:varchar(20);default:'pending'"`
	KitchenTask   *KitchenTask `json:"kitchen_task,omitempty"`
	OrderItem     *OrderItem   `json:"order_item,omitempty"`
	Table         *Table       `json:"table,omitempty"`
}
