package services

import (
"encoding/json"
"fmt"
"log"
"sort"
"time"

"restaurant-system/database"
"restaurant-system/models"
)

// DispatchResult holds information about a single dispatch operation.
type DispatchResult struct {
Task    *models.CookingTask
Message string
}

// RunAutoDispatch is the background goroutine entry point.
func RunAutoDispatch(intervalSec, windowMin int) {
ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
defer ticker.Stop()
for range ticker.C {
if err := AutoDispatch(windowMin); err != nil {
log.Printf("[dispatch] error: %v", err)
}
}
}

type pendingItem struct {
ItemID    int
OrderID   int
DishID    int
Quantity  int
IsVIP     bool
OrderedAt time.Time
TableID   int
}

// AutoDispatch aggregates pending order items within the time window and creates cooking tasks.
func AutoDispatch(windowMin int) error {
since := time.Now().Add(-time.Duration(windowMin) * time.Minute)

rows, err := database.DB.Raw(`
SELECT oi.item_id, oi.order_id, oi.dish_id, oi.quantity,
       o.is_vip, o.created_at as ordered_at, o.table_id
FROM order_item oi
JOIN `+"`order`"+` o ON o.order_id = oi.order_id
WHERE oi.status = 'pending'
  AND oi.created_at >= ?
`, since).Rows()
if err != nil {
return err
}
defer rows.Close()

type dishGroup struct {
Items    []pendingItem
TotalQty int
HasVIP   bool
TableIDs map[int]struct{}
}

groups := map[int]*dishGroup{}
for rows.Next() {
var pr pendingItem
if err := rows.Scan(&pr.ItemID, &pr.OrderID, &pr.DishID, &pr.Quantity,
&pr.IsVIP, &pr.OrderedAt, &pr.TableID); err != nil {
continue
}
g, ok := groups[pr.DishID]
if !ok {
g = &dishGroup{TableIDs: map[int]struct{}{}}
groups[pr.DishID] = g
}
g.Items = append(g.Items, pr)
g.TotalQty += pr.Quantity
if pr.IsVIP {
g.HasVIP = true
}
g.TableIDs[pr.TableID] = struct{}{}
}

for dishID, g := range groups {
itemIDs := make([]int, 0, len(g.Items))
for _, it := range g.Items {
itemIDs = append(itemIDs, it.ItemID)
}
if _, err := createTasksForDish(dishID, g.TotalQty, itemIDs, g.TableIDs, g.HasVIP); err != nil {
log.Printf("[dispatch] auto dish %d: %v", dishID, err)
}
}
return nil
}

// ManualDispatch dispatches cooking tasks for a specific order.
func ManualDispatch(orderID int) ([]DispatchResult, error) {
var order models.Order
if err := database.DB.Preload("Items").First(&order, orderID).Error; err != nil {
return nil, err
}

type group struct {
TotalQty int
ItemIDs  []int
TableIDs map[int]struct{}
HasVIP   bool
}

groups := map[int]*group{}
for _, item := range order.Items {
if item.Status != models.OrderItemStatusPending {
continue
}
g, ok := groups[item.DishID]
if !ok {
g = &group{TableIDs: map[int]struct{}{}}
groups[item.DishID] = g
}
g.TotalQty += item.Quantity
g.ItemIDs = append(g.ItemIDs, item.ItemID)
g.TableIDs[order.TableID] = struct{}{}
if order.IsVIP {
g.HasVIP = true
}
}

var results []DispatchResult
for dishID, g := range groups {
task, err := createTasksForDish(dishID, g.TotalQty, g.ItemIDs, g.TableIDs, g.HasVIP)
if err != nil {
log.Printf("[dispatch] manual dish %d: %v", dishID, err)
continue
}
if task != nil {
results = append(results, DispatchResult{Task: task, Message: "dispatched"})
}
}
return results, nil
}

func createTasksForDish(dishID, totalQty int, itemIDs []int, tableIDs map[int]struct{}, hasVIP bool) (*models.CookingTask, error) {
var dish models.Dish
if err := database.DB.First(&dish, dishID).Error; err != nil {
return nil, err
}

var recipes []models.Recipe
database.DB.Where("dish_id = ? AND is_enabled = 1", dishID).Order("portion DESC").Find(&recipes)
if len(recipes) == 0 {
return nil, nil
}

recipe := findBestRecipe(recipes, totalQty, dish.AllowCombine)
if recipe == nil {
recipe = &recipes[0]
}

chef, err := assignChef(recipe.RecipeID)
if err != nil {
return nil, err
}
if chef == nil {
return nil, nil
}

mergedJSON, _ := json.Marshal(itemIDs)
tids := make([]int, 0, len(tableIDs))
for tid := range tableIDs {
tids = append(tids, tid)
}
sort.Ints(tids)
tableJSON, _ := json.Marshal(tids)

priority := 0
if hasVIP {
priority = 10
}

task := models.CookingTask{
ChefID:       chef.ChefID,
RecipeID:     recipe.RecipeID,
DishID:       dishID,
TotalPortion: totalQty,
MergedFrom:   mergedJSON,
TableIDs:     tableJSON,
Status:       models.TaskStatusPending,
Priority:     priority,
}
if err := database.DB.Create(&task).Error; err != nil {
return nil, err
}

database.DB.Model(&models.Chef{}).Where("chef_id = ?", chef.ChefID).
UpdateColumn("current_load", chef.CurrentLoad+totalQty)

database.DB.Model(&models.OrderItem{}).Where("item_id IN ?", itemIDs).
Update("status", models.OrderItemStatusDispatched)

database.DB.Preload("Chef").Preload("Recipe").Preload("Dish").First(&task, task.TaskID)
return &task, nil
}

func findBestRecipe(recipes []models.Recipe, qty int, allowCombine bool) *models.Recipe {
// Exact match
for i := range recipes {
if recipes[i].Portion == qty {
return &recipes[i]
}
}
// Smallest recipe that still covers the total (ascending search from end)
for i := len(recipes) - 1; i >= 0; i-- {
if recipes[i].Portion >= qty {
return &recipes[i]
}
}
if allowCombine {
// Largest recipe whose portion fits within total
for i := range recipes {
if recipes[i].Portion <= qty {
return &recipes[i]
}
}
}
return nil
}

func assignChef(recipeID int) (*models.Chef, error) {
var chefs []models.Chef
err := database.DB.
Joins("JOIN chef_recipe cr ON cr.chef_id = chef.chef_id AND cr.recipe_id = ?", recipeID).
Where("chef.is_active = 1 AND chef.current_load < chef.max_load").
Order("chef.current_load ASC").
Find(&chefs).Error
if err != nil {
return nil, err
}
if len(chefs) == 0 {
return nil, nil
}
return &chefs[0], nil
}

func ReassignTask(taskID, newChefID int) (*models.CookingTask, error) {
var task models.CookingTask
if err := database.DB.First(&task, taskID).Error; err != nil {
return nil, err
}
// Release load from old chef
database.DB.Model(&models.Chef{}).Where("chef_id = ?", task.ChefID).
UpdateColumn("current_load", database.DB.Raw("GREATEST(0, current_load - ?)", task.TotalPortion))
task.ChefID = newChefID
database.DB.Save(&task)
database.DB.Model(&models.Chef{}).Where("chef_id = ?", newChefID).
UpdateColumn("current_load", database.DB.Raw("current_load + ?", task.TotalPortion))
database.DB.Preload("Chef").Preload("Recipe").Preload("Dish").First(&task, taskID)
return &task, nil
}

// ManualDispatchItem dispatches a single pending order item to a chef.
// If chefIDOverride > 0, assigns to that specific chef; otherwise auto-assigns the best available chef.
func ManualDispatchItem(orderID, itemID, chefIDOverride int) (*DispatchResult, error) {
var item models.OrderItem
if err := database.DB.Where("item_id = ? AND order_id = ?", itemID, orderID).First(&item).Error; err != nil {
return nil, fmt.Errorf("item not found: %w", err)
}
if item.Status != models.OrderItemStatusPending {
return nil, fmt.Errorf("item is not in pending status")
}
var order models.Order
if err := database.DB.First(&order, orderID).Error; err != nil {
return nil, fmt.Errorf("order not found: %w", err)
}

var recipes []models.Recipe
database.DB.Where("dish_id = ? AND is_enabled = 1", item.DishID).Order("portion DESC").Find(&recipes)
if len(recipes) == 0 {
return nil, fmt.Errorf("no recipe found for dish %d", item.DishID)
}
recipe := findBestRecipe(recipes, item.Quantity, false)
if recipe == nil {
recipe = &recipes[0]
}

var chef *models.Chef
if chefIDOverride > 0 {
var c models.Chef
if err := database.DB.First(&c, chefIDOverride).Error; err != nil {
return nil, fmt.Errorf("chef %d not found: %w", chefIDOverride, err)
}
chef = &c
} else {
var err error
chef, err = assignChef(recipe.RecipeID)
if err != nil {
return nil, err
}
if chef == nil {
return nil, fmt.Errorf("no available chef for dish %d", item.DishID)
}
}

tableJSON, _ := json.Marshal([]int{order.TableID})
mergedJSON, _ := json.Marshal([]int{item.ItemID})
priority := 0
if order.IsVIP {
priority = 10
}

task := models.CookingTask{
ChefID:       chef.ChefID,
RecipeID:     recipe.RecipeID,
DishID:       item.DishID,
TotalPortion: item.Quantity,
MergedFrom:   mergedJSON,
TableIDs:     tableJSON,
Status:       models.TaskStatusPending,
Priority:     priority,
}
if err := database.DB.Create(&task).Error; err != nil {
return nil, err
}

database.DB.Model(&models.Chef{}).Where("chef_id = ?", chef.ChefID).
UpdateColumn("current_load", chef.CurrentLoad+item.Quantity)
database.DB.Model(&models.OrderItem{}).Where("item_id = ?", item.ItemID).
Update("status", models.OrderItemStatusDispatched)

database.DB.Preload("Chef").Preload("Recipe").Preload("Dish").First(&task, task.TaskID)
return &DispatchResult{Task: &task, Message: "dispatched"}, nil
}
