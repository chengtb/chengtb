package services

import (
	"testing"

	"github.com/chengtb/restaurant-kds/internal/models"
)

func TestSplitOrderToTasks(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	orderSvc := NewOrderService(db, hub)
	kitchenSvc := NewKitchenService(db, hub, 5)

	// Create staff and order
	staff := models.Staff{Name: "Waiter1", Role: models.StaffRoleWaiter}
	db.Create(&staff)

	var table models.Table
	db.First(&table)
	var products []models.Product
	db.Where("status = ?", models.ProductStatusActive).Find(&products)

	order, _ := orderSvc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items: []CreateOrderItem{
			{ProductID: products[0].ID, Quantity: 2},
			{ProductID: products[1].ID, Quantity: 1},
		},
	})

	// Confirm order first
	orderSvc.ConfirmOrder(order.ID, staff.ID)

	// Split into tasks
	tasks, err := kitchenSvc.SplitOrderToTasks(order.ID)
	if err != nil {
		t.Fatalf("SplitOrderToTasks failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	// Verify order status changed
	var updatedOrder models.Order
	db.First(&updatedOrder, order.ID)
	if updatedOrder.Status != models.OrderStatusCooking {
		t.Errorf("Expected cooking status, got %s", updatedOrder.Status)
	}
}

func TestSplitOrderRequiresConfirmed(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	orderSvc := NewOrderService(db, hub)
	kitchenSvc := NewKitchenService(db, hub, 5)

	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	order, _ := orderSvc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})

	// Try to split without confirming - should fail
	_, err := kitchenSvc.SplitOrderToTasks(order.ID)
	if err == nil {
		t.Error("Expected error splitting unconfirmed order")
	}
}

func TestTaskMerging(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	orderSvc := NewOrderService(db, hub)
	kitchenSvc := NewKitchenService(db, hub, 5)

	staff := models.Staff{Name: "Waiter1", Role: models.StaffRoleWaiter}
	db.Create(&staff)

	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	// Create and confirm first order
	order1, _ := orderSvc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})
	orderSvc.ConfirmOrder(order1.ID, staff.ID)
	tasks1, _ := kitchenSvc.SplitOrderToTasks(order1.ID)

	// Create and confirm second order for same product (within merge window)
	order2, _ := orderSvc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 2}},
	})
	orderSvc.ConfirmOrder(order2.ID, staff.ID)
	tasks2, _ := kitchenSvc.SplitOrderToTasks(order2.ID)

	// Should be merged into the same task
	if len(tasks2) != 1 {
		t.Fatalf("Expected 1 merged task, got %d", len(tasks2))
	}

	if tasks2[0].ID != tasks1[0].ID {
		t.Error("Expected tasks to be merged into same task")
	}

	if tasks2[0].TargetQuantity != 3 {
		t.Errorf("Expected merged quantity 3, got %d", tasks2[0].TargetQuantity)
	}

	if tasks2[0].Type != models.KitchenTaskTypeMerged {
		t.Errorf("Expected merged type, got %s", tasks2[0].Type)
	}
}

func TestCompleteTask(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	orderSvc := NewOrderService(db, hub)
	kitchenSvc := NewKitchenService(db, hub, 5)

	staff := models.Staff{Name: "Waiter1", Role: models.StaffRoleWaiter}
	db.Create(&staff)

	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	order, _ := orderSvc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})
	orderSvc.ConfirmOrder(order.ID, staff.ID)
	tasks, _ := kitchenSvc.SplitOrderToTasks(order.ID)

	// Complete the task
	completed, err := kitchenSvc.CompleteTask(tasks[0].ID)
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}
	if completed.Status != models.KitchenTaskStatusCompleted {
		t.Errorf("Expected completed status, got %s", completed.Status)
	}

	// Verify voice notification created
	var notifications []models.VoiceNotification
	db.Where("task_id = ?", tasks[0].ID).Find(&notifications)
	if len(notifications) == 0 {
		t.Error("Expected voice notification to be created")
	}
}

func TestReassignTask(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	orderSvc := NewOrderService(db, hub)
	kitchenSvc := NewKitchenService(db, hub, 5)

	staff := models.Staff{Name: "Waiter1", Role: models.StaffRoleWaiter}
	db.Create(&staff)

	chef := models.Chef{Name: "Chef1", Skills: `["stir_fry"]`, Status: "available"}
	db.Create(&chef)

	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	order, _ := orderSvc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})
	orderSvc.ConfirmOrder(order.ID, staff.ID)
	tasks, _ := kitchenSvc.SplitOrderToTasks(order.ID)

	reassigned, err := kitchenSvc.ReassignTask(tasks[0].ID, chef.ID)
	if err != nil {
		t.Fatalf("ReassignTask failed: %v", err)
	}
	if *reassigned.AssignedChefID != chef.ID {
		t.Errorf("Expected chef ID %d, got %d", chef.ID, *reassigned.AssignedChefID)
	}
	if reassigned.Status != models.KitchenTaskStatusAssigned {
		t.Errorf("Expected assigned status, got %s", reassigned.Status)
	}
}

