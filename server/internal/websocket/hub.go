// Package websocket manages real-time WebSocket connections and message broadcasting.
package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a single connected WebSocket client.
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// Hub maintains the set of active clients and broadcasts messages to all of them.
// All mutations to the client map are serialised through the register/unregister
// channels to avoid data races without holding a mutex during sends.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

// NewHub creates and returns an uninitialised Hub. Call Run() in a goroutine.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run is the hub's event loop. It must be started in its own goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Slow client: drop the message and disconnect.
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// RegisterAndServe upgrades an HTTP connection to WebSocket, registers the client
// with the hub, and starts pumping outbound messages until the connection closes.
func (h *Hub) RegisterAndServe(conn *websocket.Conn) {
	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}
	h.register <- client

	// writePump forwards messages from the hub to the WebSocket connection.
	go func() {
		defer func() {
			h.unregister <- client
			conn.Close()
		}()
		for msg := range client.send {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("ws write: %v", err)
				return
			}
		}
	}()

	// readPump keeps the connection alive and handles client-side closes.
	go func() {
		defer func() {
			h.unregister <- client
			conn.Close()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
