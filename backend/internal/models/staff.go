package models

type StaffRole string

const (
	StaffRoleWaiter  StaffRole = "waiter"
	StaffRoleLeader  StaffRole = "leader"
	StaffRoleManager StaffRole = "manager"
)

type Staff struct {
	Model
	Name      string    `json:"name" gorm:"type:varchar(100);not null"`
	Role      StaffRole `json:"role" gorm:"type:varchar(20);not null"`
	Phone     string    `json:"phone" gorm:"type:varchar(20)"`
	AuthCode  string    `json:"-" gorm:"type:varchar(100);comment:authorization code for approvals"`
	RegionIDs string    `json:"region_ids" gorm:"type:json;comment:JSON array of managed region IDs"`
}
