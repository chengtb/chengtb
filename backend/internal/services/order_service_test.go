package services

import (
	"encoding/json"
	"testing"

	"github.com/chengtb/restaurant-kds/internal/models"
	ws "github.com/chengtb/restaurant-kds/internal/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	err = db.AutoMigrate(
		&models.Region{}, &models.Table{}, &models.Category{}, &models.Product{},
		&models.ProductSpec{}, &models.Recipe{}, &models.Chef{},
		&models.Order{}, &models.OrderItem{}, &models.OrderLog{},
		&models.KitchenTask{}, &models.TaskItem{}, &models.VoiceNotification{},
		&models.RefundRecord{}, &models.GiftRecord{},
		&models.Coupon{}, &models.UserCoupon{}, &models.Promotion{},
		&models.Staff{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	return db
}

func setupTestHub() *ws.Hub {
	hub := ws.NewHub()
	go hub.Run()
	return hub
}

func seedTestData(t *testing.T, db *gorm.DB) {
	// Create region
	region := models.Region{Name: "Main Hall"}
	db.Create(&region)

	// Create table
	table := models.Table{
		RegionID:    region.ID,
		TableNumber: "A1",
		Status:      models.TableStatusFree,
	}
	db.Create(&table)

	// Create category
	category := models.Category{Name: "Main Dishes", Sort: 1}
	db.Create(&category)

	// Create products
	products := []models.Product{
		{CategoryID: category.ID, Name: "Kung Pao Chicken", Type: models.ProductTypeSingle, Price: 38.0, Status: models.ProductStatusActive},
		{CategoryID: category.ID, Name: "Mapo Tofu", Type: models.ProductTypeSingle, Price: 28.0, Status: models.ProductStatusActive},
		{CategoryID: category.ID, Name: "Orange Juice", Type: models.ProductTypeDrink, Price: 15.0, Status: models.ProductStatusActive},
		{CategoryID: category.ID, Name: "Sold Out Dish", Type: models.ProductTypeSingle, Price: 50.0, Status: models.ProductStatusSoldOut},
	}
	db.Create(&products)

	// Create specs
	specs := []models.ProductSpec{
		{ProductID: products[0].ID, SpecName: "Spice Level", SpecValue: "Mild", ExtraPrice: 0},
		{ProductID: products[0].ID, SpecName: "Spice Level", SpecValue: "Extra Spicy", ExtraPrice: 2},
		{ProductID: products[2].ID, SpecName: "Size", SpecValue: "Small", ExtraPrice: 0},
		{ProductID: products[2].ID, SpecName: "Size", SpecValue: "Large", ExtraPrice: 5},
	}
	db.Create(&specs)
}

func TestCreateOrder(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	svc := NewOrderService(db, hub)

	// Get table and products
	var table models.Table
	db.First(&table)
	var products []models.Product
	db.Where("status = ?", models.ProductStatusActive).Find(&products)

	req := &CreateOrderRequest{
		TableID: table.ID,
		Remark:  "No peanuts please",
		Items: []CreateOrderItem{
			{ProductID: products[0].ID, Quantity: 2},
			{ProductID: products[1].ID, Quantity: 1},
		},
	}

	order, err := svc.CreateOrder(req)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	if order.Status != models.OrderStatusPending {
		t.Errorf("Expected status pending, got %s", order.Status)
	}
	if order.OrderSN == "" {
		t.Error("Order SN should not be empty")
	}
	if len(order.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(order.Items))
	}
	expectedTotal := products[0].Price*2 + products[1].Price*1
	if order.TotalAmount != expectedTotal {
		t.Errorf("Expected total %.2f, got %.2f", expectedTotal, order.TotalAmount)
	}

	// Verify table status changed
	var updatedTable models.Table
	db.First(&updatedTable, table.ID)
	if updatedTable.Status != models.TableStatusOccupied {
		t.Errorf("Expected table status occupied, got %s", updatedTable.Status)
	}

	// Verify order log
	var logs []models.OrderLog
	db.Where("order_id = ?", order.ID).Find(&logs)
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}
}

func TestCreateOrderWithSpecs(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	svc := NewOrderService(db, hub)

	var table models.Table
	db.First(&table)
	var product models.Product
	db.First(&product)

	specDetail := json.RawMessage(`[{"spec_id": 1, "extra_price": 0}]`)
	req := &CreateOrderRequest{
		TableID: table.ID,
		Items: []CreateOrderItem{
			{ProductID: product.ID, SpecDetail: specDetail, Quantity: 1},
		},
	}

	order, err := svc.CreateOrder(req)
	if err != nil {
		t.Fatalf("CreateOrder with specs failed: %v", err)
	}

	if len(order.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(order.Items))
	}
}

func TestCreateOrderSoldOut(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	svc := NewOrderService(db, hub)

	var table models.Table
	db.First(&table)
	var soldOutProduct models.Product
	db.Where("status = ?", models.ProductStatusSoldOut).First(&soldOutProduct)

	req := &CreateOrderRequest{
		TableID: table.ID,
		Items: []CreateOrderItem{
			{ProductID: soldOutProduct.ID, Quantity: 1},
		},
	}

	_, err := svc.CreateOrder(req)
	if err == nil {
		t.Error("Expected error for sold-out product, got nil")
	}
}

func TestConfirmOrder(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	svc := NewOrderService(db, hub)

	// Create staff
	staff := models.Staff{Name: "Waiter1", Role: models.StaffRoleWaiter}
	db.Create(&staff)

	// Create order
	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	order, _ := svc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})

	confirmed, err := svc.ConfirmOrder(order.ID, staff.ID)
	if err != nil {
		t.Fatalf("ConfirmOrder failed: %v", err)
	}
	if confirmed.Status != models.OrderStatusConfirmed {
		t.Errorf("Expected confirmed status, got %s", confirmed.Status)
	}
}

func TestConfirmOrderWrongStatus(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	svc := NewOrderService(db, hub)
	staff := models.Staff{Name: "Waiter1", Role: models.StaffRoleWaiter}
	db.Create(&staff)

	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	order, _ := svc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})

	// Confirm once
	svc.ConfirmOrder(order.ID, staff.ID)

	// Try to confirm again
	_, err := svc.ConfirmOrder(order.ID, staff.ID)
	if err == nil {
		t.Error("Expected error confirming already-confirmed order")
	}
}

func TestGetOrder(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	svc := NewOrderService(db, hub)

	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	created, _ := svc.CreateOrder(&CreateOrderRequest{
		TableID: table.ID,
		Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
	})

	order, err := svc.GetOrder(created.ID)
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if order.ID != created.ID {
		t.Errorf("Expected order ID %d, got %d", created.ID, order.ID)
	}
}

func TestListOrders(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	seedTestData(t, db)

	svc := NewOrderService(db, hub)

	var table models.Table
	db.First(&table)
	var product models.Product
	db.Where("status = ?", models.ProductStatusActive).First(&product)

	// Create 3 orders
	for i := 0; i < 3; i++ {
		svc.CreateOrder(&CreateOrderRequest{
			TableID: table.ID,
			Items:   []CreateOrderItem{{ProductID: product.ID, Quantity: 1}},
		})
	}

	orders, total, err := svc.ListOrders("", 1, 10)
	if err != nil {
		t.Fatalf("ListOrders failed: %v", err)
	}
	if total != 3 {
		t.Errorf("Expected 3 orders, got %d", total)
	}
	if len(orders) != 3 {
		t.Errorf("Expected 3 orders in list, got %d", len(orders))
	}

	// Filter by status
	orders, total, _ = svc.ListOrders("pending", 1, 10)
	if total != 3 {
		t.Errorf("Expected 3 pending orders, got %d", total)
	}
}

func TestGenerateOrderSN(t *testing.T) {
	db := setupTestDB(t)
	hub := setupTestHub()
	svc := NewOrderService(db, hub)

	sn1 := svc.GenerateOrderSN()
	sn2 := svc.GenerateOrderSN()

	if sn1 == "" || sn2 == "" {
		t.Error("Order SN should not be empty")
	}
	if len(sn1) < 10 {
		t.Errorf("Order SN too short: %s", sn1)
	}
}
