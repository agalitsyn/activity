package model

import "errors"

type Agent struct {
	ID string
}

var ErrAgentNotFound = errors.New("agent not found")

type AgentRepository interface {
	FetchAgent(id string) (Agent, error)
	CreateAgent(c Agent) error
	DeleteAgent(id string) error
	FetchAgents() ([]Agent, error)
}
