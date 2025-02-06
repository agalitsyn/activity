package mem

import (
	"sync"

	"github.com/agalitsyn/activity/internal/model"
)

type ClientStorage struct {
	mu   sync.RWMutex
	data map[string]model.Client
}

func NewClientStorage() *ClientStorage {
	return &ClientStorage{
		data: make(map[string]model.Client),
	}
}

func (s *ClientStorage) CreateClient(client model.Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[client.ID] = client
	return nil
}

func (s *ClientStorage) FetchClient(id string) (model.Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.data[id]
	if !ok {
		return model.Client{}, model.ErrClientNotFound
	}
	return client, nil
}

func (s *ClientStorage) DeleteClient(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, id)
	return nil
}

func (s *ClientStorage) FetchClients() ([]model.Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]model.Client, 0, len(s.data))
	for _, c := range s.data {
		clients = append(clients, c)
	}
	return clients, nil
}
