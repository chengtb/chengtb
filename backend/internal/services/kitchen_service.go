package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chengtb/restaurant-kds/internal/models"
	ws "github.com/chengtb/restaurant-kds/internal/websocket"
	"gorm.io/gorm"
)

type KitchenService struct {
	DB                 *gorm.DB
	Hub                *ws.Hub
	MergeWindowMinutes int
}

func NewKitchenService(db *gorm.DB, hub *ws.Hub, mergeWindowMinutes int) *KitchenService {
	return &KitchenService{DB: db, Hub: hub, MergeWindowMinutes: mergeWindowMinutes}
}

// SplitOrderToTasks splits a confirmed order into kitchen tasks.
// Packages are split into individual items, each product+spec combination
// maps to a recipe, and tasks can be merged with existing tasks for the
// same recipe within the merge window.
func (s *KitchenService) SplitOrderToTasks(orderID uint) ([]models.KitchenTask, error) {
	var order models.Order
	if err := s.DB.Preload("Items.Product").First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.Status != models.OrderStatusConfirmed {
		return nil, fmt.Errorf("order must be confirmed before splitting")
	}

	var tasks []models.KitchenTask

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range order.Items {
			if item.Status != models.OrderItemStatusPending {
				continue
			}

			if item.Product != nil && item.Product.Type == models.ProductTypePackage {
				subTasks, err := s.splitPackageItem(tx, &order, &item)
				if err != nil {
					return err
				}
				tasks = append(tasks, subTasks...)
			} else {
				task, err := s.createOrMergeTask(tx, &order, &item)
				if err != nil {
					return err
				}
				tasks = append(tasks, *task)
			}

			tx.Model(&item).Update("status", models.OrderItemStatusCooking)
		}

		tx.Model(&order).Update("status", models.OrderStatusCooking)

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.Hub.BroadcastToRole(ws.ClientTypeKDS, ws.Message{
		Type: "new_tasks",
		Data: tasks,
	})

	return tasks, nil
}

// PackageSubItem represents a sub-item in a package product
type PackageSubItem struct {
	ProductID  uint            `json:"product_id"`
	Quantity   int             `json:"quantity"`
	SpecDetail json.RawMessage `json:"spec_detail"`
}

func (s *KitchenService) splitPackageItem(tx *gorm.DB, order *models.Order, item *models.OrderItem) ([]models.KitchenTask, error) {
	var subItems []PackageSubItem
	if item.Product.PackageItems != "" {
		if err := json.Unmarshal([]byte(item.Product.PackageItems), &subItems); err != nil {
			// If package items are not defined, treat as single item
			task, err := s.createOrMergeTask(tx, order, item)
			if err != nil {
				return nil, err
			}
			return []models.KitchenTask{*task}, nil
		}
	}

	var tasks []models.KitchenTask
	for _, sub := range subItems {
		var product models.Product
		if err := tx.First(&product, sub.ProductID).Error; err != nil {
			continue
		}

		specStr := ""
		if sub.SpecDetail != nil {
			specStr = string(sub.SpecDetail)
		}

		subOrderItem := &models.OrderItem{
			OrderID:    order.ID,
			ProductID:  sub.ProductID,
			SpecDetail: specStr,
			UnitPrice:  0, // package sub-items don't have individual prices
			Quantity:   sub.Quantity * item.Quantity,
			Status:     models.OrderItemStatusCooking,
		}
		tx.Create(subOrderItem)

		task, err := s.createOrMergeTask(tx, order, subOrderItem)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

func (s *KitchenService) createOrMergeTask(tx *gorm.DB, order *models.Order, item *models.OrderItem) (*models.KitchenTask, error) {
	recipe, err := s.findRecipe(tx, item)
	if err != nil {
		return nil, err
	}

	// Check for existing mergeable task within the time window
	mergeWindow := time.Now().Add(-time.Duration(s.MergeWindowMinutes) * time.Minute)
	var existingTask models.KitchenTask
	canMerge := tx.Where(
		"recipe_id = ? AND status = ? AND created_at > ?",
		recipe.ID, models.KitchenTaskStatusPending, mergeWindow,
	).First(&existingTask).Error == nil

	if canMerge {
		existingTask.TargetQuantity += item.Quantity
		existingTask.Type = models.KitchenTaskTypeMerged
		tx.Save(&existingTask)

		tx.Create(&models.TaskItem{
			KitchenTaskID: existingTask.ID,
			OrderItemID:   item.ID,
			TableID:       order.TableID,
			Status:        "pending",
		})

		return &existingTask, nil
	}

	now := time.Now()
	batchNo := fmt.Sprintf("BATCH%s%04d", now.Format("20060102150405"), now.Nanosecond()%10000)
	task := &models.KitchenTask{
		Type:           models.KitchenTaskTypeNormal,
		RecipeID:       recipe.ID,
		BatchNo:        batchNo,
		TargetQuantity: item.Quantity,
		Status:         models.KitchenTaskStatusPending,
	}
	if err := tx.Create(task).Error; err != nil {
		return nil, err
	}

	tx.Create(&models.TaskItem{
		KitchenTaskID: task.ID,
		OrderItemID:   item.ID,
		TableID:       order.TableID,
		Status:        "pending",
	})

	s.autoAssignChef(tx, task, recipe)

	return task, nil
}

func (s *KitchenService) findRecipe(tx *gorm.DB, item *models.OrderItem) (*models.Recipe, error) {
	var recipe models.Recipe
	query := tx.Where("product_id = ?", item.ProductID)
	if item.SpecDetail != "" && item.SpecDetail != "null" {
		query = query.Where("spec_combination = ?", item.SpecDetail)
	}
	if err := query.First(&recipe).Error; err != nil {
		// Create a default recipe if none exists
		recipe = models.Recipe{
			ProductID:        item.ProductID,
			SpecCombination:  item.SpecDetail,
			DefaultBatchSize: 1,
		}
		tx.Create(&recipe)
	}
	return &recipe, nil
}

func (s *KitchenService) autoAssignChef(tx *gorm.DB, task *models.KitchenTask, recipe *models.Recipe) {
	if recipe.ChefRequiredSkill == "" {
		return
	}

	var chefs []models.Chef
	tx.Where("status = ?", "available").Find(&chefs)

	for _, chef := range chefs {
		var skills []string
		if err := json.Unmarshal([]byte(chef.Skills), &skills); err != nil {
			continue
		}
		for _, skill := range skills {
			if skill == recipe.ChefRequiredSkill {
				task.AssignedChefID = &chef.ID
				task.Status = models.KitchenTaskStatusAssigned
				tx.Save(task)

				s.Hub.BroadcastToRole(ws.ClientTypeChef, ws.Message{
					Type: "task_assigned",
					Data: task,
				})
				return
			}
		}
	}
}

// CompleteTask marks a kitchen task as completed
func (s *KitchenService) CompleteTask(taskID uint) (*models.KitchenTask, error) {
	var task models.KitchenTask
	if err := s.DB.Preload("TaskItems").First(&task, taskID).Error; err != nil {
		return nil, fmt.Errorf("task not found")
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		task.Status = models.KitchenTaskStatusCompleted
		if err := tx.Save(&task).Error; err != nil {
			return err
		}

		for _, ti := range task.TaskItems {
			tx.Model(&ti).Update("status", "completed")
			tx.Model(&models.OrderItem{}).Where("id = ?", ti.OrderItemID).Update("status", models.OrderItemStatusCooked)

			tx.Create(&models.VoiceNotification{
				TableID: ti.TableID,
				TaskID:  task.ID,
				Content: fmt.Sprintf("桌号 %d 的菜品已完成", ti.TableID),
				Status:  "pending",
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.Hub.BroadcastToRole(ws.ClientTypeWaiter, ws.Message{
		Type: "dish_ready",
		Data: task,
	})

	return &task, nil
}

// ReassignTask reassigns a task to a different chef
func (s *KitchenService) ReassignTask(taskID uint, newChefID uint) (*models.KitchenTask, error) {
	var task models.KitchenTask
	if err := s.DB.First(&task, taskID).Error; err != nil {
		return nil, fmt.Errorf("task not found")
	}
	if task.Status == models.KitchenTaskStatusCompleted || task.Status == models.KitchenTaskStatusCancelled {
		return nil, fmt.Errorf("cannot reassign completed or cancelled task")
	}

	task.AssignedChefID = &newChefID
	task.Status = models.KitchenTaskStatusAssigned
	s.DB.Save(&task)

	s.Hub.BroadcastToRole(ws.ClientTypeChef, ws.Message{
		Type: "task_reassigned",
		Data: task,
	})

	return &task, nil
}

// ListTasks lists kitchen tasks with optional filters
func (s *KitchenService) ListTasks(status string, chefID uint, page, pageSize int) ([]models.KitchenTask, int64, error) {
	var tasks []models.KitchenTask
	var total int64

	query := s.DB.Model(&models.KitchenTask{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if chefID > 0 {
		query = query.Where("assigned_chef_id = ?", chefID)
	}

	query.Count(&total)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	err := query.Preload("Recipe.Product").Preload("TaskItems.Table").Preload("Chef").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at ASC").Find(&tasks).Error

	return tasks, total, err
}
