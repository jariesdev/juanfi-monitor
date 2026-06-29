package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins to match the Python CORS configuration.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler returns a Gin handler that upgrades the HTTP connection to WebSocket
// and hands off the connection to the Hub for lifetime management.
func Handler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("ws upgrade: %v", err)
			return
		}
		hub.RegisterAndServe(conn)
	}
}
