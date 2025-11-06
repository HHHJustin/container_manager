package containers

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockProvider 以記憶體模擬容器生命週期，便於測試與無 Docker 環境。
type MockProvider struct {
	mu         sync.RWMutex
	containers map[string]Container
}

func NewMockProvider() *MockProvider {
	return &MockProvider{containers: make(map[string]Container)}
}

func (m *MockProvider) Create(opts CreateOptions) (Container, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := uuid.NewString()
	c := Container{
		ID:        id,
		Name:      opts.Name,
		Image:     opts.Image,
		CreatedAt: time.Now().Unix(),
		Status:    "created",
	}
	m.containers[id] = c
	return c, nil
}

func (m *MockProvider) Start(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.containers[id]
	if !ok {
		return ErrNotFound
	}
	c.Status = "running"
	m.containers[id] = c
	return nil
}

func (m *MockProvider) Stop(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.containers[id]
	if !ok {
		return ErrNotFound
	}
	c.Status = "stopped"
	m.containers[id] = c
	return nil
}

func (m *MockProvider) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.containers[id]; !ok {
		return ErrNotFound
	}
	delete(m.containers, id)
	return nil
}
