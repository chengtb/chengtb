package models

type ProductSpec struct {
	Model
	ProductID  uint     `json:"product_id" gorm:"not null;index"`
	SpecName   string   `json:"spec_name" gorm:"type:varchar(100);not null"`
	SpecValue  string   `json:"spec_value" gorm:"type:varchar(100);not null"`
	ExtraPrice float64  `json:"extra_price" gorm:"type:decimal(10,2);default:0"`
	Product    *Product `json:"product,omitempty"`
}
