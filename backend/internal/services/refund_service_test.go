package services

import (
	"testing"

	"github.com/chengtb/restaurant-kds/internal/models"
)

func TestRefundNotMade(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	// Create waiter
	waiter := models.Staff{Name: "Waiter1", Role: models.StaffRoleWaiter}
	db.Create(&waiter)

	// Create order with pending item
	orderSvc := NewOrderService(db, hub)
	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	order, _ := orderSvc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})

	refundSvc := NewRefundService(db)
	record, err := refundSvc.CreateRefund(&RefundRequest{
		OrderItemID:  order.Items[0].ID,
		RefundReason: "Customer changed mind",
		OperatorID:   waiter.ID,
	})
	if err != nil {
		t.Fatalf("Refund not-made failed: %v", err)
	}
	if record.RefundType != models.RefundTypeNotMade {
		t.Errorf("Expected refund type not_made, got %s", record.RefundType)
	}
	if record.Status != models.RefundStatusApproved {
		t.Errorf("Expected approved status, got %s", record.Status)
	}

	// Verify item status
	var item models.OrderItem
	db.First(&item, order.Items[0].ID)
	if item.Status != models.OrderItemStatusRefunded {
		t.Errorf("Expected refunded status, got %s", item.Status)
	}
}

func TestRefundMadeNotServedRequiresLeader(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	waiter := models.Staff{Name: "Waiter1", Role: models.StaffRoleWaiter}
	db.Create(&waiter)
	leader := models.Staff{Name: "Leader1", Role: models.StaffRoleLeader, AuthCode: "LEADER123"}
	db.Create(&leader)

	orderSvc := NewOrderService(db, hub)
	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	order, _ := orderSvc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})

	// Set item to cooked status
	db.Model(&models.OrderItem{}).Where("id = ?", order.Items[0].ID).Update("status", models.OrderItemStatusCooked)

	refundSvc := NewRefundService(db)

	// Without approval code - should fail
	_, err := refundSvc.CreateRefund(&RefundRequest{
		OrderItemID:  order.Items[0].ID,
		RefundReason: "Quality issue",
		OperatorID:   waiter.ID,
	})
	if err == nil {
		t.Error("Expected error without leader approval code")
	}

	// With wrong code - should fail
	_, err = refundSvc.CreateRefund(&RefundRequest{
		OrderItemID:  order.Items[0].ID,
		RefundReason: "Quality issue",
		OperatorID:   waiter.ID,
		ApprovalCode: "WRONG",
	})
	if err == nil {
		t.Error("Expected error with wrong approval code")
	}

	// With correct leader code - should succeed
	record, err := refundSvc.CreateRefund(&RefundRequest{
		OrderItemID:  order.Items[0].ID,
		RefundReason: "Quality issue",
		OperatorID:   waiter.ID,
		ApprovalCode: "LEADER123",
	})
	if err != nil {
		t.Fatalf("Refund with leader code failed: %v", err)
	}
	if record.RefundType != models.RefundTypeMadeNotServed {
		t.Errorf("Expected refund type made_not_served, got %s", record.RefundType)
	}
}

func TestRefundServedRequiresManager(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	waiter := models.Staff{Name: "Waiter1", Role: models.StaffRoleWaiter}
	db.Create(&waiter)
	manager := models.Staff{Name: "Manager1", Role: models.StaffRoleManager, AuthCode: "MGR456"}
	db.Create(&manager)

	orderSvc := NewOrderService(db, hub)
	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	order, _ := orderSvc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})

	// Set item to served status
	db.Model(&models.OrderItem{}).Where("id = ?", order.Items[0].ID).Update("status", models.OrderItemStatusServed)

	refundSvc := NewRefundService(db)

	// Without approval code - should fail
	_, err := refundSvc.CreateRefund(&RefundRequest{
		OrderItemID:  order.Items[0].ID,
		RefundReason: "Wrong dish served",
		OperatorID:   waiter.ID,
	})
	if err == nil {
		t.Error("Expected error without manager approval code")
	}

	// With manager code - should succeed
	record, err := refundSvc.CreateRefund(&RefundRequest{
		OrderItemID:  order.Items[0].ID,
		RefundReason: "Wrong dish served",
		OperatorID:   waiter.ID,
		ApprovalCode: "MGR456",
	})
	if err != nil {
		t.Fatalf("Refund with manager code failed: %v", err)
	}
	if record.RefundType != models.RefundTypeServed {
		t.Errorf("Expected refund type served, got %s", record.RefundType)
	}
}
