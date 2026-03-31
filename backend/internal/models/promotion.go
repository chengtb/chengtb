package models

type PromotionType string

const (
	PromotionTypeHalfPrice    PromotionType = "half_price"
	PromotionTypeSpecialPrice PromotionType = "special_price"
	PromotionTypeMemberPrice  PromotionType = "member_price"
)

type Promotion struct {
	Model
	Name   string        `json:"name" gorm:"type:varchar(200);not null"`
	Type   PromotionType `json:"type" gorm:"type:varchar(30);not null"`
	Rules  string        `json:"rules" gorm:"type:json;comment:promotion rules JSON"`
	Status string        `json:"status" gorm:"type:varchar(20);default:'active'"`
}
