package ws

import (
	"encoding/json"
	"log"
	"sync"

	"travel-collab/backend/internal/models"
)

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[*Client]bool)}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[c.TripID] == nil {
		h.rooms[c.TripID] = make(map[*Client]bool)
	}
	h.rooms[c.TripID][c] = true
	log.Printf("ws: client registered trip=%s user=%s", c.TripID, c.UserID)
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.rooms[c.TripID]
	if clients == nil {
		return
	}
	if _, ok := clients[c]; ok {
		delete(clients, c)
		close(c.Send)
	}
	if len(clients) == 0 {
		delete(h.rooms, c.TripID)
	}
	log.Printf("ws: client unregistered trip=%s user=%s", c.TripID, c.UserID)
}

func (h *Hub) Broadcast(tripID string, event models.WSEvent) {
	bytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("ws marshal error: %v", err)
		return
	}
	h.BroadcastBytes(tripID, bytes)
}

func (h *Hub) BroadcastBytes(tripID string, message []byte) {
	h.mu.RLock()
	clients := h.rooms[tripID]
	list := make([]*Client, 0, len(clients))
	for c := range clients {
		list = append(list, c)
	}
	h.mu.RUnlock()

	for _, c := range list {
		select {
		case c.Send <- message:
		default:
			h.Unregister(c)
		}
	}
}
