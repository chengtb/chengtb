package services

import (
	"encoding/json"

	"restaurant-system/database"
	"restaurant-system/models"
)

func GetOrCreateCartSession(tableID int) (*models.CartSession, error) {
	var session models.CartSession
	err := database.DB.Where("table_id = ? AND status = ?", tableID, models.CartSessionActive).
		Preload("Items.Dish").
		First(&session).Error
	if err == nil {
		return &session, nil
	}

	session = models.CartSession{
		TableID: tableID,
		Status:  models.CartSessionActive,
	}
	if err := database.DB.Create(&session).Error; err != nil {
		return nil, err
	}

	// Update table status to ordering
	database.DB.Model(&models.Table{}).Where("table_id = ?", tableID).
		Update("status", models.TableStatusOrdering)

	return &session, nil
}

func GetCartSession(sessionID int) (*models.CartSession, error) {
	var session models.CartSession
	err := database.DB.Where("session_id = ?", sessionID).
		Preload("Items.Dish").
		Preload("Table").
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func AddItemToCart(sessionID, dishID, quantity int, note, addedBy string) (*models.SessionItem, error) {
	// Check if dish already exists in session
	var existing models.SessionItem
	err := database.DB.Where("session_id = ? AND dish_id = ?", sessionID, dishID).First(&existing).Error
	if err == nil {
		existing.Quantity += quantity
		if err := database.DB.Save(&existing).Error; err != nil {
			return nil, err
		}
		database.DB.Preload("Dish").First(&existing, existing.ItemID)
		return &existing, nil
	}

	item := models.SessionItem{
		SessionID: sessionID,
		DishID:    dishID,
		Quantity:  quantity,
		Note:      note,
		AddedBy:   addedBy,
	}
	if err := database.DB.Create(&item).Error; err != nil {
		return nil, err
	}
	database.DB.Preload("Dish").First(&item, item.ItemID)
	return &item, nil
}

func UpdateCartItem(sessionID, itemID, quantity int) (*models.SessionItem, error) {
	var item models.SessionItem
	if err := database.DB.Where("item_id = ? AND session_id = ?", itemID, sessionID).First(&item).Error; err != nil {
		return nil, err
	}
	if quantity <= 0 {
		err := database.DB.Delete(&item).Error
		return nil, err
	}
	item.Quantity = quantity
	if err := database.DB.Save(&item).Error; err != nil {
		return nil, err
	}
	database.DB.Preload("Dish").First(&item, item.ItemID)
	return &item, nil
}

func RemoveCartItem(sessionID, itemID int) error {
	return database.DB.Where("item_id = ? AND session_id = ?", itemID, sessionID).
		Delete(&models.SessionItem{}).Error
}

func BroadcastCartUpdate(sessionID int, session *models.CartSession) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    "cart_update",
		"session": session,
	})
	if err != nil {
		return
	}
	Hub.BroadcastToSession(sessionID, data)
}
