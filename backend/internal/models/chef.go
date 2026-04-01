package models

type Chef struct {
	Model
	Name          string `json:"name" gorm:"type:varchar(100);not null"`
	Skills        string `json:"skills" gorm:"type:json;comment:JSON array of skill tags"`
	WorkstationID string `json:"workstation_id" gorm:"type:varchar(50)"`
	Status        string `json:"status" gorm:"type:varchar(20);default:'available'"`
}
