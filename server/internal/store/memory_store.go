package store

import (
	"sync"

	"github.com/alexsharav/layout-phone-backend/internal/models"
)

type MemoryStore struct {
	RoomsByCode map[string]*models.Room
	Mu          sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		RoomsByCode: make(map[string]*models.Room),
	}
}
