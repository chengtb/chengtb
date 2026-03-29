package services

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	SessionID int
	Conn      *websocket.Conn
	Send      chan []byte
}

type WSHub struct {
	mu       sync.RWMutex
	sessions map[int][]*WSClient
}

var Hub = &WSHub{
	sessions: make(map[int][]*WSClient),
}

func (h *WSHub) Register(client *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[client.SessionID] = append(h.sessions[client.SessionID], client)
}

func (h *WSHub) Unregister(client *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.sessions[client.SessionID]
	for i, c := range clients {
		if c == client {
			h.sessions[client.SessionID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	if len(h.sessions[client.SessionID]) == 0 {
		delete(h.sessions, client.SessionID)
	}
	close(client.Send)
}

func (h *WSHub) BroadcastToSession(sessionID int, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.sessions[sessionID] {
		select {
		case client.Send <- msg:
		default:
			// drop message if channel is full
		}
	}
}

func (h *WSHub) WritePump(client *WSClient) {
	defer client.Conn.Close()
	for msg := range client.Send {
		if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (h *WSHub) ReadPump(client *WSClient) {
	defer func() {
		h.Unregister(client)
		client.Conn.Close()
	}()
	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
