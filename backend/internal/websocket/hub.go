package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

type ClientType string

const (
	ClientTypeCustomer ClientType = "customer"
	ClientTypeWaiter   ClientType = "waiter"
	ClientTypeChef     ClientType = "chef"
	ClientTypeKDS      ClientType = "kds"
)

type Message struct {
	Type    string      `json:"type"`
	TableID uint        `json:"table_id,omitempty"`
	Data    interface{} `json:"data"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]bool
	tables  map[uint]map[*Client]bool       // tableID -> clients
	roles   map[ClientType]map[*Client]bool // role -> clients

	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *BroadcastMessage
}

type BroadcastMessage struct {
	TableID uint       // 0 means broadcast by role
	Role    ClientType // used when TableID is 0
	Message Message
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		tables:     make(map[uint]map[*Client]bool),
		roles:      make(map[ClientType]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *BroadcastMessage, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			if client.TableID > 0 {
				if h.tables[client.TableID] == nil {
					h.tables[client.TableID] = make(map[*Client]bool)
				}
				h.tables[client.TableID][client] = true
			}
			if h.roles[client.Type] == nil {
				h.roles[client.Type] = make(map[*Client]bool)
			}
			h.roles[client.Type][client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if client.TableID > 0 {
					delete(h.tables[client.TableID], client)
				}
				if h.roles[client.Type] != nil {
					delete(h.roles[client.Type], client)
				}
				close(client.Send)
			}
			h.mu.Unlock()

		case bm := <-h.Broadcast:
			data, err := json.Marshal(bm.Message)
			if err != nil {
				log.Printf("Failed to marshal broadcast message: %v", err)
				continue
			}
			h.mu.RLock()
			if bm.TableID > 0 {
				for client := range h.tables[bm.TableID] {
					select {
					case client.Send <- data:
					default:
						go func(c *Client) { h.Unregister <- c }(client)
					}
				}
			}
			if bm.Role != "" {
				for client := range h.roles[bm.Role] {
					select {
					case client.Send <- data:
					default:
						go func(c *Client) { h.Unregister <- c }(client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToTable sends a message to all clients connected to a specific table
func (h *Hub) BroadcastToTable(tableID uint, msg Message) {
	h.Broadcast <- &BroadcastMessage{
		TableID: tableID,
		Message: msg,
	}
}

// BroadcastToRole sends a message to all clients of a specific role
func (h *Hub) BroadcastToRole(role ClientType, msg Message) {
	h.Broadcast <- &BroadcastMessage{
		Role:    role,
		Message: msg,
	}
}
