package model

import "errors"

type Agent struct {
	ID               string
	ActiveApp        string
	ActiveAppContext string
}

var ErrAgentNotFound = errors.New("agent not found")

type AgentRepository interface {
	FetchAgent(id string) (Agent, error)
	CreateAgent(agent Agent) error
	DeleteAgent(id string) error
	FetchAgents() ([]Agent, error)
	UpdateAgent(agent Agent) error
}
