package services

import (
	"fmt"

	"github.com/chengtb/restaurant-kds/internal/models"
	"gorm.io/gorm"
)

type GiftService struct {
	DB *gorm.DB
}

func NewGiftService(db *gorm.DB) *GiftService {
	return &GiftService{DB: db}
}

// GiftRequest holds the payload for creating a gift
type GiftRequest struct {
	OrderID      uint                    `json:"order_id" binding:"required"`
	ProductID    uint                    `json:"product_id" binding:"required"`
	Quantity     int                     `json:"quantity" binding:"required,min=1"`
	TriggerScene models.GiftTriggerScene `json:"trigger_scene" binding:"required"`
	OperatorID   uint                    `json:"operator_id"`
	ApprovalCode string                  `json:"approval_code"`
	Remark       string                  `json:"remark"`
}

// CreateGift handles gift creation based on trigger scene:
//   - compensation: requires leader approval
//   - crm: can be auto-triggered by system or manually by waiter
//   - marketing: auto-approved by system rules
func (s *GiftService) CreateGift(req *GiftRequest) (*models.GiftRecord, error) {
	var order models.Order
	if err := s.DB.First(&order, req.OrderID).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}

	var product models.Product
	if err := s.DB.First(&product, req.ProductID).Error; err != nil {
		return nil, fmt.Errorf("product not found")
	}

	record := &models.GiftRecord{
		OrderID:      req.OrderID,
		ProductID:    req.ProductID,
		Quantity:     req.Quantity,
		TriggerScene: req.TriggerScene,
		Remark:       req.Remark,
	}

	if req.OperatorID > 0 {
		record.OperatorID = &req.OperatorID
	}

	switch req.TriggerScene {
	case models.GiftSceneCompensation:
		if req.ApprovalCode == "" {
			record.ApprovalStatus = models.GiftApprovalPending
		} else {
			var leader models.Staff
			if err := s.DB.Where("role IN (?, ?) AND auth_code = ?", models.StaffRoleLeader, models.StaffRoleManager, req.ApprovalCode).First(&leader).Error; err != nil {
				return nil, fmt.Errorf("invalid approval code")
			}
			record.ApprovalStatus = models.GiftApprovalApproved
		}

	case models.GiftSceneCRM:
		record.ApprovalStatus = models.GiftApprovalApproved

	case models.GiftSceneMarketing:
		record.ApprovalStatus = models.GiftApprovalAuto

	default:
		return nil, fmt.Errorf("invalid trigger scene")
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}

		if record.ApprovalStatus == models.GiftApprovalApproved || record.ApprovalStatus == models.GiftApprovalAuto {
			tx.Create(&models.OrderItem{
				OrderID:   req.OrderID,
				ProductID: req.ProductID,
				UnitPrice: 0,
				Quantity:  req.Quantity,
				Status:    models.OrderItemStatusPending,
				Remark:    fmt.Sprintf("[赠送-%s] %s", req.TriggerScene, req.Remark),
			})
		}

		tx.Create(&models.OrderLog{
			OrderID:      req.OrderID,
			OperatorID:   req.OperatorID,
			OperatorType: "staff",
			Action:       "gift",
			Content:      fmt.Sprintf("Gift: %s x%d (scene: %s)", product.Name, req.Quantity, req.TriggerScene),
		})

		return nil
	})

	return record, err
}

// ListGifts lists gift records
func (s *GiftService) ListGifts(page, pageSize int) ([]models.GiftRecord, int64, error) {
	var records []models.GiftRecord
	var total int64

	s.DB.Model(&models.GiftRecord{}).Count(&total)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	err := s.DB.Preload("Order").Preload("Product").Preload("Operator").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&records).Error

	return records, total, err
}
