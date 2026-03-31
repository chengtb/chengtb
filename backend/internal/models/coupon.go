package models

import "time"

type CouponType string

const (
	CouponTypeFullReduction CouponType = "full_reduction"
	CouponTypeDiscount      CouponType = "discount"
	CouponTypeVoucher       CouponType = "voucher"
)

type Coupon struct {
	Model
	Name      string     `json:"name" gorm:"type:varchar(200);not null"`
	Type      CouponType `json:"type" gorm:"type:varchar(30);not null"`
	Value     float64    `json:"value" gorm:"type:decimal(10,2);not null;comment:discount value or percentage"`
	Condition float64    `json:"condition" gorm:"type:decimal(10,2);default:0;comment:minimum order amount"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	Status    string     `json:"status" gorm:"type:varchar(20);default:'active'"`
}
