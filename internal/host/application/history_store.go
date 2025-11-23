package application

import (
	"sync"

	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
)

type MemoryHistoryStore struct {
	mu   sync.RWMutex
	data map[string][]ai_provider.Message
}

func NewMemoryHistoryStore() *MemoryHistoryStore {
	return &MemoryHistoryStore{data: make(map[string][]ai_provider.Message)}
}

func (s *MemoryHistoryStore) Get(convID string) []ai_provider.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if msgs, ok := s.data[convID]; ok {
		return append([]ai_provider.Message(nil), msgs...)
	}
	return []ai_provider.Message{}
}

func (s *MemoryHistoryStore) Set(convID string, msgs []ai_provider.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[convID] = append([]ai_provider.Message(nil), msgs...)
}
