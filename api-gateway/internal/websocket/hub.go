package websocket

import (
	"log"
	"sync"
	"time"

	gwebsocket "github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
	sendBufferSize = 16
)

type Client struct {
	hub    *Hub
	conn   *gwebsocket.Conn
	send   chan []byte
	userID string
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

func (h *Hub) Register(userID string, conn *gwebsocket.Conn) *Client {
	client := &Client{
		hub:    h,
		conn:   conn,
		send:   make(chan []byte, sendBufferSize),
		userID: userID,
	}

	h.mu.Lock()
	if oldClient, exists := h.clients[userID]; exists {
		close(oldClient.send)
	}
	h.clients[userID] = client
	h.mu.Unlock()

	go client.writePump()
	go client.readPump()

	return client
}

func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	currentClient, exists := h.clients[client.userID]
	if !exists || currentClient != client {
		return
	}

	delete(h.clients, client.userID)
	close(client.send)
}

func (h *Hub) Send(userID string, payload []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, exists := h.clients[userID]
	if !exists {
		return false
	}

	select {
	case client.send <- payload:
		return true
	default:
		log.Printf("ws: send buffer full for user=%s, dropping notification", userID)
		return false
	}
}

func (client *Client) readPump() {
	defer func() {
		client.hub.unregister(client)
		_ = client.conn.Close()
	}()

	client.conn.SetReadLimit(maxMessageSize)

	if err := client.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("ws: failed to set read deadline for user=%s: %v", client.userID, err)
		return
	}

	client.conn.SetPongHandler(func(string) error {
		return client.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		if _, _, err := client.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = client.conn.Close()
	}()

	for {
		select {
		case payload, ok := <-client.send:
			if err := client.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("ws: failed to set write deadline for user=%s: %v", client.userID, err)
				return
			}

			if !ok {
				_ = client.conn.WriteMessage(gwebsocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WriteMessage(gwebsocket.TextMessage, payload); err != nil {
				return
			}

		case <-ticker.C:
			if err := client.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("ws: failed to set ping write deadline for user=%s: %v", client.userID, err)
				return
			}

			if err := client.conn.WriteMessage(gwebsocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
