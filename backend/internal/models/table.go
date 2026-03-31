package models

type TableStatus string

const (
	TableStatusFree     TableStatus = "free"
	TableStatusOccupied TableStatus = "occupied"
	TableStatusReserved TableStatus = "reserved"
)

type Table struct {
	Model
	RegionID    uint        `json:"region_id" gorm:"not null;index"`
	TableNumber string      `json:"table_number" gorm:"type:varchar(50);not null;uniqueIndex"`
	QRCode      string      `json:"qrcode" gorm:"type:varchar(500)"`
	Status      TableStatus `json:"status" gorm:"type:varchar(20);default:'free'"`
	Capacity    int         `json:"capacity" gorm:"default:4"`
	Region      *Region     `json:"region,omitempty"`
}
