package websocket

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

func HandleWebSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}

		tableIDStr := c.Query("table_id")
		clientType := ClientType(c.Query("type"))

		var tableID uint
		if tableIDStr != "" {
			id, err := strconv.ParseUint(tableIDStr, 10, 32)
			if err == nil {
				tableID = uint(id)
			}
		}

		if clientType == "" {
			clientType = ClientTypeCustomer
		}

		client := &Client{
			Hub:     hub,
			Conn:    conn,
			Send:    make(chan []byte, 256),
			TableID: tableID,
			Type:    clientType,
		}

		hub.Register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}
