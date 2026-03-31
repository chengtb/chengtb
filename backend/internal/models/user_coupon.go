package models

type UserCouponStatus string

const (
	UserCouponStatusUnused  UserCouponStatus = "unused"
	UserCouponStatusUsed    UserCouponStatus = "used"
	UserCouponStatusExpired UserCouponStatus = "expired"
)

type UserCoupon struct {
	Model
	UserID   string           `json:"user_id" gorm:"type:varchar(100);not null;index"`
	CouponID uint             `json:"coupon_id" gorm:"not null;index"`
	Status   UserCouponStatus `json:"status" gorm:"type:varchar(20);default:'unused'"`
	Coupon   *Coupon          `json:"coupon,omitempty"`
}
