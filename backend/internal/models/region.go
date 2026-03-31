package models

type Region struct {
	Model
	Name   string  `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	Tables []Table `json:"tables,omitempty" gorm:"foreignKey:RegionID"`
}
