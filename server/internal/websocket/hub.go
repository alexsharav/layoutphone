package websocket

import "sync"

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[string]*Client),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[client.RoomCode] == nil {
		h.rooms[client.RoomCode] = make(map[string]*Client)
	}
	h.rooms[client.RoomCode][client.PlayerID] = client
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room := h.rooms[client.RoomCode]
	if room == nil {
		return
	}
	delete(room, client.PlayerID)
	if len(room) == 0 {
		delete(h.rooms, client.RoomCode)
	}
}

func (h *Hub) Broadcast(roomCode string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.rooms[roomCode] {
		client.Send(event)
	}
}

func (h *Hub) SendToPlayer(roomCode string, playerID string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client := h.rooms[roomCode][playerID]; client != nil {
		client.Send(event)
	}
}
