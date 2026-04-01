package services

import (
	"fmt"

	"github.com/chengtb/restaurant-kds/internal/config"
	"github.com/chengtb/restaurant-kds/internal/models"
	"gorm.io/gorm"
)

type RefundService struct {
	DB                *gorm.DB
	RefundPermissions *config.RefundPermissionConfig
}

func NewRefundService(db *gorm.DB, perms *config.RefundPermissionConfig) *RefundService {
	return &RefundService{DB: db, RefundPermissions: perms}
}

// RefundRequest holds the payload for creating a refund
type RefundRequest struct {
	OrderItemID  uint   `json:"order_item_id" binding:"required"`
	RefundReason string `json:"refund_reason" binding:"required"`
	OperatorID   uint   `json:"operator_id" binding:"required"`
	ApprovalCode string `json:"approval_code"`
}

// CreateRefund handles refund with permission checks:
//   - not_made: waiter can directly refund
//   - made_not_served: requires leader approval code
//   - served: requires manager approval code
func (s *RefundService) CreateRefund(req *RefundRequest) (*models.RefundRecord, error) {
	var orderItem models.OrderItem
	if err := s.DB.First(&orderItem, req.OrderItemID).Error; err != nil {
		return nil, fmt.Errorf("order item not found")
	}

	var refundType models.RefundType
	switch orderItem.Status {
	case models.OrderItemStatusPending, models.OrderItemStatusCooking:
		refundType = models.RefundTypeNotMade
	case models.OrderItemStatusCooked:
		refundType = models.RefundTypeMadeNotServed
	case models.OrderItemStatusServed:
		refundType = models.RefundTypeServed
	default:
		return nil, fmt.Errorf("item cannot be refunded in current status")
	}

	if err := s.checkRefundPermission(refundType, req.OperatorID, req.ApprovalCode); err != nil {
		return nil, err
	}

	record := &models.RefundRecord{
		OrderItemID:  req.OrderItemID,
		RefundReason: req.RefundReason,
		RefundType:   refundType,
		Status:       models.RefundStatusApproved,
		ApproverID:   &req.OperatorID,
		ApprovalCode: req.ApprovalCode,
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}

		tx.Model(&orderItem).Update("status", models.OrderItemStatusRefunded)

		tx.Create(&models.OrderLog{
			OrderID:      orderItem.OrderID,
			OperatorID:   req.OperatorID,
			OperatorType: "staff",
			Action:       "refund",
			Content:      fmt.Sprintf("Refund: %s (type: %s)", req.RefundReason, refundType),
		})

		return nil
	})

	return record, err
}

func (s *RefundService) checkRefundPermission(refundType models.RefundType, operatorID uint, approvalCode string) error {
	switch refundType {
	case models.RefundTypeNotMade:
		// Configured role: s.RefundPermissions.NotMade (default "waiter")
		// Any valid staff member can refund items that have not been made.
		var staff models.Staff
		if err := s.DB.First(&staff, operatorID).Error; err != nil {
			return fmt.Errorf("operator not found")
		}
		return nil

	case models.RefundTypeMadeNotServed:
		requiredRole := models.StaffRole(s.RefundPermissions.MadeNotServed)
		if approvalCode == "" {
			return fmt.Errorf("%s approval code required for made-but-not-served refund", requiredRole)
		}
		var approver models.Staff
		if err := s.DB.Where("role = ? AND auth_code = ?", requiredRole, approvalCode).First(&approver).Error; err != nil {
			return fmt.Errorf("invalid %s approval code", requiredRole)
		}
		return nil

	case models.RefundTypeServed:
		requiredRole := models.StaffRole(s.RefundPermissions.Served)
		if approvalCode == "" {
			return fmt.Errorf("%s approval code required for served-dish refund", requiredRole)
		}
		var approver models.Staff
		if err := s.DB.Where("role = ? AND auth_code = ?", requiredRole, approvalCode).First(&approver).Error; err != nil {
			return fmt.Errorf("invalid %s approval code", requiredRole)
		}
		return nil
	}

	return fmt.Errorf("unknown refund type")
}

// ListRefunds lists refund records
func (s *RefundService) ListRefunds(page, pageSize int) ([]models.RefundRecord, int64, error) {
	var records []models.RefundRecord
	var total int64

	s.DB.Model(&models.RefundRecord{}).Count(&total)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	err := s.DB.Preload("OrderItem.Product").Preload("Approver").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&records).Error

	return records, total, err
}
