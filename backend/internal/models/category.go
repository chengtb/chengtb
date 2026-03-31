package models

type Category struct {
	Model
	Name     string    `json:"name" gorm:"type:varchar(100);not null"`
	Sort     int       `json:"sort" gorm:"default:0"`
	Products []Product `json:"products,omitempty" gorm:"foreignKey:CategoryID"`
}
