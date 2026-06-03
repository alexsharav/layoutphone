package websocket

import (
	"sync"

	gorilla "github.com/gorilla/websocket"
)

type Client struct {
	RoomCode string
	PlayerID string
	conn     *gorilla.Conn
	hub      *Hub
	mu       sync.Mutex
}

func NewClient(roomCode string, playerID string, conn *gorilla.Conn, hub *Hub) *Client {
	return &Client{
		RoomCode: roomCode,
		PlayerID: playerID,
		conn:     conn,
		hub:      hub,
	}
}

func (c *Client) ReadLoop() {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (c *Client) Send(event Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.conn.WriteJSON(event)
}
