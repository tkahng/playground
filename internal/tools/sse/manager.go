package sse

import (
	"context"
	"sync"
)

type Manager interface {
	Clients() []Client
	RegisterClient(context.Context, context.CancelFunc, Client)
	UnregisterClient(Client)
	Run(context.Context)

	Send(clientKey string)
}

type manager struct {
	mu         *sync.RWMutex
	clients    map[Client]context.CancelFunc
	register   chan regreq
	unregister chan regreq
}
type regreq struct {
	context context.Context
	cancel  context.CancelFunc
	client  Client
	done    chan struct{}
}
