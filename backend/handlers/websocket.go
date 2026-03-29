package handlers

import (
	"net/http"
	"strconv"

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
