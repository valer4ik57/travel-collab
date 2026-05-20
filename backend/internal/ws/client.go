package ws

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"travel-collab/backend/internal/models"
	"travel-collab/backend/internal/repository"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	Hub    *Hub
	Store  *repository.Store
	Conn   *websocket.Conn
	Send   chan []byte
	TripID string
	UserID string
}

type inboundMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type sendMessagePayload struct {
	Text string `json:"text"`
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		_ = c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws read error: %v", err)
			}
			break
		}
		c.handleMessage(message)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(raw []byte) {
	var msg inboundMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return
	}
	switch msg.Type {
	case "SEND_MESSAGE":
		role, err := c.Store.GetTripMemberRole(context.Background(), c.TripID, c.UserID)
		if err != nil || (role != "owner" && role != "editor") {
			return
		}
		var payload sendMessagePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		text := strings.TrimSpace(payload.Text)
		if text == "" || len([]rune(text)) > 1000 {
			return
		}
		saved, err := c.Store.SaveMessage(context.Background(), c.TripID, c.UserID, text)
		if err != nil {
			log.Printf("save message error: %v", err)
			return
		}
		c.Hub.Broadcast(c.TripID, models.WSEvent{Type: "MESSAGE_SENT", Payload: saved})
	}
}
