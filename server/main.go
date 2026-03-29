package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// ─── Models ────────────────────────────────────────────────────────────────────

type Table struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Number   string `json:"number" gorm:"uniqueIndex"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Status   string `json:"status" gorm:"default:'available'"` // available/occupied/reserved
}

type Category struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	Icon      string `json:"icon"`
}

type Dish struct {
	ID          uint     `json:"id" gorm:"primaryKey"`
	CategoryID  uint     `json:"category_id"`
	Category    Category `json:"category" gorm:"foreignKey:CategoryID"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	ImageURL    string   `json:"image_url"`
	Available   bool     `json:"available" gorm:"default:true"`
	SortOrder   int      `json:"sort_order"`
}

type Order struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	TableID     uint        `json:"table_id"`
	Table       Table       `json:"table" gorm:"foreignKey:TableID"`
	Status      string      `json:"status" gorm:"default:'pending'"` // pending/confirmed/cooking/served/paid/cancelled
	TotalAmount float64     `json:"total_amount"`
	Note        string      `json:"note"`
	Items       []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	OrderID   uint    `json:"order_id"`
	DishID    uint    `json:"dish_id"`
	DishName  string  `json:"dish_name"`
	DishPrice float64 `json:"dish_price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
	Status    string  `json:"status" gorm:"default:'pending'"` // pending/cooking/done
	Note      string  `json:"note"`
}

type CookingTask struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	OrderItemID uint       `json:"order_item_id"`
	OrderID     uint       `json:"order_id"`
	DishName    string     `json:"dish_name"`
	Quantity    int        `json:"quantity"`
	Status      string     `json:"status" gorm:"default:'pending'"` // pending/cooking/done
	TableNumber string     `json:"table_number"`
	AssignedAt  time.Time  `json:"assigned_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// ─── Globals ────────────────────────────────────────────────────────────────────

var db *gorm.DB

// ─── Database Init + Seed ───────────────────────────────────────────────────────

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("restaurant.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	db.AutoMigrate(&Table{}, &Category{}, &Dish{}, &Order{}, &OrderItem{}, &CookingTask{})

	seedData()
}

func seedData() {
	// Tables
	var tableCount int64
	db.Model(&Table{}).Count(&tableCount)
	if tableCount == 0 {
		tables := []Table{
			{Number: "A1", Name: "A区1号桌", Capacity: 4, Status: "available"},
			{Number: "A2", Name: "A区2号桌", Capacity: 4, Status: "available"},
			{Number: "A3", Name: "A区3号桌", Capacity: 6, Status: "available"},
			{Number: "A4", Name: "A区4号桌", Capacity: 6, Status: "available"},
			{Number: "B1", Name: "B区1号桌", Capacity: 2, Status: "available"},
			{Number: "B2", Name: "B区2号桌", Capacity: 2, Status: "available"},
			{Number: "B3", Name: "B区3号桌", Capacity: 4, Status: "available"},
			{Number: "B4", Name: "B区4号桌", Capacity: 8, Status: "available"},
			{Number: "C1", Name: "C区1号桌", Capacity: 4, Status: "available"},
			{Number: "C2", Name: "C区2号桌", Capacity: 4, Status: "available"},
			{Number: "C3", Name: "C区3号桌", Capacity: 6, Status: "available"},
			{Number: "C4", Name: "C区4号桌", Capacity: 10, Status: "available"},
		}
		db.Create(&tables)
	}

	// Categories
	var catCount int64
	db.Model(&Category{}).Count(&catCount)
	if catCount == 0 {
		categories := []Category{
			{Name: "热菜", SortOrder: 1, Icon: "🍳"},
			{Name: "凉菜", SortOrder: 2, Icon: "🥗"},
			{Name: "汤品", SortOrder: 3, Icon: "🍲"},
			{Name: "主食", SortOrder: 4, Icon: "🍚"},
			{Name: "饮品", SortOrder: 5, Icon: "🥤"},
		}
		db.Create(&categories)

		// Dishes
		var cats []Category
		db.Find(&cats)
		catMap := map[string]uint{}
		for _, c := range cats {
			catMap[c.Name] = c.ID
		}

		dishes := []Dish{
			// 热菜
			{CategoryID: catMap["热菜"], Name: "宫保鸡丁", Description: "经典川菜，鸡肉嫩滑，花生香脆", Price: 38.0, Available: true, SortOrder: 1},
			{CategoryID: catMap["热菜"], Name: "麻婆豆腐", Description: "豆腐嫩滑，麻辣鲜香", Price: 28.0, Available: true, SortOrder: 2},
			{CategoryID: catMap["热菜"], Name: "红烧肉", Description: "肥而不腻，入口即化", Price: 58.0, Available: true, SortOrder: 3},
			{CategoryID: catMap["热菜"], Name: "鱼香肉丝", Description: "酸甜辣鲜，下饭神器", Price: 32.0, Available: true, SortOrder: 4},
			{CategoryID: catMap["热菜"], Name: "糖醋排骨", Description: "酸甜口味，外酥里嫩", Price: 68.0, Available: true, SortOrder: 5},
			{CategoryID: catMap["热菜"], Name: "清炒时蔬", Description: "新鲜时令蔬菜，清淡爽口", Price: 22.0, Available: true, SortOrder: 6},
			{CategoryID: catMap["热菜"], Name: "干锅花椰菜", Description: "香辣入味，鲜嫩可口", Price: 35.0, Available: true, SortOrder: 7},
			{CategoryID: catMap["热菜"], Name: "剁椒鱼头", Description: "湖南特色，鲜辣可口", Price: 88.0, Available: true, SortOrder: 8},
			// 凉菜
			{CategoryID: catMap["凉菜"], Name: "拍黄瓜", Description: "清脆爽口，蒜香浓郁", Price: 18.0, Available: true, SortOrder: 1},
			{CategoryID: catMap["凉菜"], Name: "凉拌木耳", Description: "爽滑Q弹，酸辣开胃", Price: 22.0, Available: true, SortOrder: 2},
			{CategoryID: catMap["凉菜"], Name: "夫妻肺片", Description: "麻辣鲜香，肉质软烂", Price: 38.0, Available: true, SortOrder: 3},
			{CategoryID: catMap["凉菜"], Name: "皮蛋豆腐", Description: "嫩滑豆腐配皮蛋，鲜美可口", Price: 25.0, Available: true, SortOrder: 4},
			// 汤品
			{CategoryID: catMap["汤品"], Name: "番茄蛋花汤", Description: "酸甜可口，营养丰富", Price: 18.0, Available: true, SortOrder: 1},
			{CategoryID: catMap["汤品"], Name: "酸辣汤", Description: "酸辣开胃，暖胃驱寒", Price: 22.0, Available: true, SortOrder: 2},
			{CategoryID: catMap["汤品"], Name: "老火靓汤", Description: "慢火熬制，营养滋补", Price: 48.0, Available: true, SortOrder: 3},
			// 主食
			{CategoryID: catMap["主食"], Name: "白米饭", Description: "东北大米，颗粒饱满", Price: 3.0, Available: true, SortOrder: 1},
			{CategoryID: catMap["主食"], Name: "蛋炒饭", Description: "金黄蛋炒饭，香气扑鼻", Price: 18.0, Available: true, SortOrder: 2},
			{CategoryID: catMap["主食"], Name: "葱油饼", Description: "层次分明，葱香浓郁", Price: 12.0, Available: true, SortOrder: 3},
			// 饮品
			{CategoryID: catMap["饮品"], Name: "可乐", Description: "冰镇可口可乐 330ml", Price: 8.0, Available: true, SortOrder: 1},
			{CategoryID: catMap["饮品"], Name: "菊花茶", Description: "清热去火，甘甜清香", Price: 12.0, Available: true, SortOrder: 2},
			{CategoryID: catMap["饮品"], Name: "鲜榨橙汁", Description: "新鲜橙子现榨，维C丰富", Price: 18.0, Available: true, SortOrder: 3},
		}
		db.Create(&dishes)
	}
}

// ─── Handlers ───────────────────────────────────────────────────────────────────

// Tables
func getTables(c *gin.Context) {
	var tables []Table
	db.Find(&tables)
	c.JSON(http.StatusOK, tables)
}

func getTable(c *gin.Context) {
	id := c.Param("id")
	var table Table
	if err := db.First(&table, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "桌台不存在"})
		return
	}
	c.JSON(http.StatusOK, table)
}

func updateTableStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Model(&Table{}).Where("id = ?", id).Update("status", req.Status)
	c.JSON(http.StatusOK, gin.H{"message": "状态已更新"})
}

// Categories
func getCategories(c *gin.Context) {
	var categories []Category
	db.Order("sort_order").Find(&categories)
	c.JSON(http.StatusOK, categories)
}

func createCategory(c *gin.Context) {
	var cat Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Create(&cat)
	c.JSON(http.StatusCreated, cat)
}

func updateCategory(c *gin.Context) {
	id := c.Param("id")
	var cat Category
	if err := db.First(&cat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "分类不存在"})
		return
	}
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Save(&cat)
	c.JSON(http.StatusOK, cat)
}

func deleteCategory(c *gin.Context) {
	id := c.Param("id")
	db.Delete(&Category{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// Dishes
func getDishes(c *gin.Context) {
	var dishes []Dish
	query := db.Preload("Category").Order("sort_order")
	if catID := c.Query("category_id"); catID != "" {
		query = query.Where("category_id = ?", catID)
	}
	query.Find(&dishes)
	c.JSON(http.StatusOK, dishes)
}

func createDish(c *gin.Context) {
	var dish Dish
	if err := c.ShouldBindJSON(&dish); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Create(&dish)
	c.JSON(http.StatusCreated, dish)
}

func updateDish(c *gin.Context) {
	id := c.Param("id")
	var dish Dish
	if err := db.First(&dish, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "菜品不存在"})
		return
	}
	if err := c.ShouldBindJSON(&dish); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Save(&dish)
	c.JSON(http.StatusOK, dish)
}

func deleteDish(c *gin.Context) {
	id := c.Param("id")
	db.Delete(&Dish{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// Orders
type CreateOrderRequest struct {
	TableID uint   `json:"table_id"`
	Note    string `json:"note"`
	Items   []struct {
		DishID   uint   `json:"dish_id"`
		Quantity int    `json:"quantity"`
		Note     string `json:"note"`
	} `json:"items"`
}

func createOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var table Table
	if err := db.First(&table, req.TableID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "桌台不存在"})
		return
	}

	order := Order{
		TableID: req.TableID,
		Status:  "pending",
		Note:    req.Note,
	}

	var total float64
	var orderItems []OrderItem
	var cookingTasks []CookingTask

	for _, item := range req.Items {
		var dish Dish
		if err := db.First(&dish, item.DishID).Error; err != nil {
			continue
		}
		subtotal := dish.Price * float64(item.Quantity)
		total += subtotal
		oi := OrderItem{
			DishID:    dish.ID,
			DishName:  dish.Name,
			DishPrice: dish.Price,
			Quantity:  item.Quantity,
			Subtotal:  subtotal,
			Status:    "pending",
			Note:      item.Note,
		}
		orderItems = append(orderItems, oi)
	}
	order.TotalAmount = total
	order.Items = orderItems

	db.Create(&order)

	// Create cooking tasks
	for _, oi := range order.Items {
		task := CookingTask{
			OrderItemID: oi.ID,
			OrderID:     order.ID,
			DishName:    oi.DishName,
			Quantity:    oi.Quantity,
			Status:      "pending",
			TableNumber: table.Number,
			AssignedAt:  time.Now(),
		}
		cookingTasks = append(cookingTasks, task)
	}
	if len(cookingTasks) > 0 {
		db.Create(&cookingTasks)
	}

	// Set table to occupied
	db.Model(&Table{}).Where("id = ?", req.TableID).Update("status", "occupied")

	db.Preload("Items").Preload("Table").First(&order, order.ID)
	c.JSON(http.StatusCreated, order)
}

func getOrders(c *gin.Context) {
	var orders []Order
	query := db.Preload("Items").Preload("Table").Order("created_at desc")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if tableID := c.Query("table_id"); tableID != "" {
		query = query.Where("table_id = ?", tableID)
	}
	query.Find(&orders)
	c.JSON(http.StatusOK, orders)
}

func getOrder(c *gin.Context) {
	id := c.Param("id")
	var order Order
	if err := db.Preload("Items").Preload("Table").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func updateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status         string  `json:"status"`
		DiscountAmount float64 `json:"discount_amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order Order
	if err := db.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	updates := map[string]interface{}{"status": req.Status}
	if req.DiscountAmount > 0 {
		newTotal := order.TotalAmount - req.DiscountAmount
		if newTotal < 0 {
			newTotal = 0
		}
		updates["total_amount"] = newTotal
	}
	db.Model(&order).Updates(updates)

	// When paid, set table back to available
	if req.Status == "paid" {
		db.Model(&Table{}).Where("id = ?", order.TableID).Update("status", "available")
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态已更新"})
}

func addOrderItems(c *gin.Context) {
	id := c.Param("id")
	var order Order
	if err := db.Preload("Table").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	var req struct {
		Items []struct {
			DishID   uint   `json:"dish_id"`
			Quantity int    `json:"quantity"`
			Note     string `json:"note"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var addedTotal float64
	for _, item := range req.Items {
		var dish Dish
		if err := db.First(&dish, item.DishID).Error; err != nil {
			continue
		}
		subtotal := dish.Price * float64(item.Quantity)
		addedTotal += subtotal
		oi := OrderItem{
			OrderID:   order.ID,
			DishID:    dish.ID,
			DishName:  dish.Name,
			DishPrice: dish.Price,
			Quantity:  item.Quantity,
			Subtotal:  subtotal,
			Status:    "pending",
			Note:      item.Note,
		}
		db.Create(&oi)

		task := CookingTask{
			OrderItemID: oi.ID,
			OrderID:     order.ID,
			DishName:    oi.DishName,
			Quantity:    oi.Quantity,
			Status:      "pending",
			TableNumber: order.Table.Number,
			AssignedAt:  time.Now(),
		}
		db.Create(&task)
	}

	db.Model(&order).Update("total_amount", order.TotalAmount+addedTotal)
	db.Preload("Items").Preload("Table").First(&order, id)
	c.JSON(http.StatusOK, order)
}

// Cooking Tasks
func getCookingTasks(c *gin.Context) {
	var tasks []CookingTask
	query := db.Order("assigned_at desc")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	query.Find(&tasks)
	c.JSON(http.StatusOK, tasks)
}

func updateCookingTaskStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{"status": req.Status}
	if req.Status == "done" {
		now := time.Now()
		updates["completed_at"] = now
	}
	db.Model(&CookingTask{}).Where("id = ?", id).Updates(updates)
	c.JSON(http.StatusOK, gin.H{"message": "状态已更新"})
}

// QR Code
func getQRCode(c *gin.Context) {
	tableID := c.Param("tableId")
	tableIDInt, err := strconv.Atoi(tableID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的桌台ID"})
		return
	}

	var table Table
	if err := db.First(&table, tableIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "桌台不存在"})
		return
	}

	// URL that customer will scan - points to mobile app
	url := fmt.Sprintf("http://localhost:5173/table/%d", tableIDInt)

	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成二维码失败"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=table-%s-qrcode.png", table.Number))
	c.Data(http.StatusOK, "image/png", png)
}

// ─── Main ───────────────────────────────────────────────────────────────────────

func main() {
	initDB()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length", "Content-Disposition"},
	}))

	api := r.Group("/api")
	{
		// Tables
		api.GET("/tables", getTables)
		api.GET("/tables/:id", getTable)
		api.PUT("/tables/:id/status", updateTableStatus)

		// Categories
		api.GET("/categories", getCategories)
		api.POST("/categories", createCategory)
		api.PUT("/categories/:id", updateCategory)
		api.DELETE("/categories/:id", deleteCategory)

		// Dishes
		api.GET("/dishes", getDishes)
		api.POST("/dishes", createDish)
		api.PUT("/dishes/:id", updateDish)
		api.DELETE("/dishes/:id", deleteDish)

		// Orders
		api.POST("/orders", createOrder)
		api.GET("/orders", getOrders)
		api.GET("/orders/:id", getOrder)
		api.PUT("/orders/:id/status", updateOrderStatus)
		api.POST("/orders/:id/items", addOrderItems)

		// Cooking Tasks
		api.GET("/cooking-tasks", getCookingTasks)
		api.PUT("/cooking-tasks/:id/status", updateCookingTaskStatus)

		// QR Code
		api.GET("/qrcode/:tableId", getQRCode)
	}

	r.Run(":8080")
}
