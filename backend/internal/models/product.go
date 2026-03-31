package models

type ProductType string

const (
	ProductTypeSingle  ProductType = "single"
	ProductTypeDrink   ProductType = "drink"
	ProductTypePackage ProductType = "package"
)

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
	ProductStatusSoldOut  ProductStatus = "sold_out"
)

type Product struct {
	Model
	CategoryID  uint          `json:"category_id" gorm:"not null;index"`
	Name        string        `json:"name" gorm:"type:varchar(200);not null"`
	Type        ProductType   `json:"type" gorm:"type:varchar(20);not null"`
	Price       float64       `json:"price" gorm:"type:decimal(10,2);not null"`
	Image       string        `json:"image" gorm:"type:varchar(500)"`
	Description string        `json:"description" gorm:"type:text"`
	Status      ProductStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	Category    *Category     `json:"category,omitempty"`
	Specs       []ProductSpec `json:"specs,omitempty" gorm:"foreignKey:ProductID"`
	// For packages: sub-items stored as JSON
	PackageItems string `json:"package_items,omitempty" gorm:"type:text"`
}
