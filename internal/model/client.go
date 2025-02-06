package model

import "errors"

type Client struct {
	ID string
}

var ErrClientNotFound = errors.New("client not found")

type ClientRepository interface {
	FetchClient(id string) (Client, error)
	CreateClient(c Client) error
	DeleteClient(id string) error
	FetchClients() ([]Client, error)
}
