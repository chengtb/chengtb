package models

type Recipe struct {
	Model
	ProductID         uint     `json:"product_id" gorm:"not null;index"`
	SpecCombination   string   `json:"spec_combination" gorm:"type:json"`
	ChefRequiredSkill string   `json:"chef_required_skill" gorm:"type:varchar(100)"`
	CookingDuration   int      `json:"cooking_duration" gorm:"default:0;comment:minutes"`
	DefaultBatchSize  int      `json:"default_batch_size" gorm:"default:1"`
	Product           *Product `json:"product,omitempty"`
}
