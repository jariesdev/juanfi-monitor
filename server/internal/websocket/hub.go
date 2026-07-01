// Package websocket manages real-time WebSocket connections and message broadcasting.
package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait is the time allowed to write a message to a peer.
	writeWait = 10 * time.Second
	// pongWait is how long we wait for a pong before considering the peer dead.
	pongWait = 60 * time.Second
	// pingPeriod is how often we ping the peer. Kept well under pongWait and under
	// typical NAT/firewall/proxy idle timeouts (~30-60s) so the socket never goes
	// idle long enough to be reaped by an intermediary.
	pingPeriod = 25 * time.Second
	// maxMessageSize caps inbound frames; clients are not expected to send data.
	maxMessageSize = 4096
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

	// writePump forwards hub messages to the connection and sends periodic pings
	// to keep the socket alive through idle-killing intermediaries (NAT, firewalls,
	// load balancers) and to detect a half-open connection.
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			h.unregister <- client
			conn.Close()
		}()
		for {
			select {
			case msg, ok := <-client.send:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					// The hub closed the channel: tell the peer and stop.
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					log.Printf("ws write: %v", err)
					return
				}
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	// readPump drains inbound frames (clients aren't expected to send any) and
	// enforces the pong deadline: each pong — browsers auto-reply to our pings —
	// extends it, so a peer that stops responding is dropped after pongWait.
	go func() {
		defer func() {
			h.unregister <- client
			conn.Close()
		}()
		conn.SetReadLimit(maxMessageSize)
		conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
