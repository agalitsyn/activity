package mem

import (
	"sync"

	"github.com/agalitsyn/activity/internal/model"
)

type AgentStorage struct {
	mu   sync.RWMutex
	data map[string]model.Agent
}

func NewAgentStorage() *AgentStorage {
	return &AgentStorage{
		data: make(map[string]model.Agent),
	}
}

func (s *AgentStorage) CreateAgent(agent model.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[agent.ID] = agent
	return nil
}

func (s *AgentStorage) FetchAgent(id string) (model.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agent, ok := s.data[id]
	if !ok {
		return model.Agent{}, model.ErrAgentNotFound
	}
	return agent, nil
}

func (s *AgentStorage) DeleteAgent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, id)
	return nil
}

func (s *AgentStorage) FetchAgents() ([]model.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agents := make([]model.Agent, 0, len(s.data))
	for _, c := range s.data {
		agents = append(agents, c)
	}
	return agents, nil
}

func (s *AgentStorage) UpdateAgent(agent model.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[agent.ID]
	if !ok {
		return model.ErrAgentNotFound
	}

	s.data[agent.ID] = agent
	return nil
}
