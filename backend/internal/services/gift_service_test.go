package services

import (
	"testing"

	"github.com/chengtb/restaurant-kds/internal/models"
)

func TestGiftCompensation(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	leader := models.Staff{Name: "Leader1", Role: models.StaffRoleLeader, AuthCode: "LEADER123"}
	db.Create(&leader)
	waiter := models.Staff{Name: "Waiter1", Role: models.StaffRoleWaiter}
	db.Create(&waiter)

	orderSvc := NewOrderService(db, hub)
	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	order, _ := orderSvc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})

	giftSvc := NewGiftService(db)

	// Without approval code - should be pending
	record, err := giftSvc.CreateGift(&GiftRequest{
		OrderID:      order.ID,
		ProductID:    product.ID,
		Quantity:     1,
		TriggerScene: models.GiftSceneCompensation,
		OperatorID:   waiter.ID,
		Remark:       "Slow service",
	})
	if err != nil {
		t.Fatalf("CreateGift compensation failed: %v", err)
	}
	if record.ApprovalStatus != models.GiftApprovalPending {
		t.Errorf("Expected pending status without approval, got %s", record.ApprovalStatus)
	}

	// With leader approval code - should be approved
	record2, err := giftSvc.CreateGift(&GiftRequest{
		OrderID:      order.ID,
		ProductID:    product.ID,
		Quantity:     1,
		TriggerScene: models.GiftSceneCompensation,
		OperatorID:   waiter.ID,
		ApprovalCode: "LEADER123",
		Remark:       "Quality issue compensation",
	})
	if err != nil {
		t.Fatalf("CreateGift with approval failed: %v", err)
	}
	if record2.ApprovalStatus != models.GiftApprovalApproved {
		t.Errorf("Expected approved status, got %s", record2.ApprovalStatus)
	}
}

func TestGiftCRM(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	waiter := models.Staff{Name: "Waiter1", Role: models.StaffRoleWaiter}
	db.Create(&waiter)

	orderSvc := NewOrderService(db, hub)
	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	order, _ := orderSvc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})

	giftSvc := NewGiftService(db)
	record, err := giftSvc.CreateGift(&GiftRequest{
		OrderID:      order.ID,
		ProductID:    product.ID,
		Quantity:     1,
		TriggerScene: models.GiftSceneCRM,
		OperatorID:   waiter.ID,
		Remark:       "Regular customer birthday",
	})
	if err != nil {
		t.Fatalf("CreateGift CRM failed: %v", err)
	}
	if record.ApprovalStatus != models.GiftApprovalApproved {
		t.Errorf("Expected auto-approved for CRM, got %s", record.ApprovalStatus)
	}
}

func TestGiftMarketing(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	orderSvc := NewOrderService(db, hub)
	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	order, _ := orderSvc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})

	giftSvc := NewGiftService(db)
	record, err := giftSvc.CreateGift(&GiftRequest{
		OrderID:      order.ID,
		ProductID:    product.ID,
		Quantity:     1,
		TriggerScene: models.GiftSceneMarketing,
		Remark:       "New dish tasting event",
	})
	if err != nil {
		t.Fatalf("CreateGift marketing failed: %v", err)
	}
	if record.ApprovalStatus != models.GiftApprovalAuto {
		t.Errorf("Expected auto status for marketing, got %s", record.ApprovalStatus)
	}
}
