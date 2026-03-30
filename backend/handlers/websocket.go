package handlers

import (
	"net/http"
	"strconv"

	"restaurant-system/database"
	"restaurant-system/models"
	"restaurant-system/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins for development
	},
}

// CartWebSocket WS /ws/cart/:sessionId
func CartWebSocket(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("sessionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &services.WSClient{
		SessionID: sessionID,
		Conn:      conn,
		Send:      make(chan []byte, 256),
	}

	services.Hub.Register(client)

	go services.Hub.WritePump(client)
	services.Hub.ReadPump(client) // blocks until disconnected
}

// WaiterWebSocket WS /ws/waiter/:waiterId
func WaiterWebSocket(c *gin.Context) {
	waiterID, err := strconv.Atoi(c.Param("waiterId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid waiter id"})
		return
	}

	var waiter models.Waiter
	if err := database.DB.First(&waiter, waiterID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "waiter not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	areaID := 0
	if waiter.AreaID != nil {
		areaID = *waiter.AreaID
	}

	client := &services.WaiterWSClient{
		WaiterID: waiterID,
		AreaID:   areaID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	// Mark waiter as online
	database.DB.Model(&models.Waiter{}).Where("waiter_id = ?", waiterID).Update("status", models.WaiterStatusOnline)

	services.WaiterHub.Register(client)
	go services.WaiterHub.WritePump(client)
	services.WaiterHub.ReadPump(client) // blocks until disconnected

	// Mark waiter as offline when disconnected
	database.DB.Model(&models.Waiter{}).Where("waiter_id = ?", waiterID).Update("status", models.WaiterStatusOffline)
}
