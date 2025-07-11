package sse

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

type Manager interface {
	Clients() []Client
	RegisterClient(context.Context, context.CancelFunc, Client)
	UnregisterClient(Client)
	Run(context.Context)

	Send(clientId string, data any) error
	SendAll(data any) error
}

type clientContext struct {
	id     string
	client Client
	cancel context.CancelFunc
}

func (c *clientContext) Client() Client {
	return c.client
}
func (c *clientContext) Cancel() context.CancelFunc {
	return c.cancel
}

func (c *clientContext) ID() string {
	return c.id
}

type ClientContext interface {
	ID() string
	Client() Client
	Cancel() context.CancelFunc
}

type manager struct {
	logger     *slog.Logger
	mu         *sync.RWMutex
	clients    map[string]ClientContext
	register   chan regreq
	unregister chan regreq
}

// Send implements Manager.
func (m *manager) Send(clientId string, data any) error {
	if c, ok := m.clients[clientId]; ok {
		return c.Client().Write(Message{Data: data})
	} else {
		m.logger.Warn("client not found", "id", clientId)
	}
	return nil
}

// SendAll implements Manager.
func (m *manager) SendAll(data any) error {
	var errs []error
	for _, c := range m.clients {
		if err := c.Client().Write(Message{Data: data}); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

type regreq struct {
	context context.Context
	cancel  context.CancelFunc
	client  Client
	done    chan struct{}
}

func NewManager(logger *slog.Logger) Manager {
	return &manager{
		mu:         &sync.RWMutex{},
		logger:     logger,
		clients:    make(map[string]ClientContext),
		register:   make(chan regreq),
		unregister: make(chan regreq),
	}
}

// Clients returns the currently managed Clients.
func (m *manager) Clients() []Client {
	res := []Client{}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, c := range m.clients {
		res = append(res, c.Client())
	}
	return res
}

// RegisterClient adds the Client to the Manager's store.
func (m *manager) RegisterClient(ctx context.Context, cf context.CancelFunc, c Client) {
	done := make(chan struct{})
	rr := regreq{
		context: ctx,
		cancel:  cf,
		client:  c,
		done:    done,
	}
	m.register <- rr
	<-done
}

// UnregisterClient removes the Client from the Manager's store.
func (m *manager) UnregisterClient(c Client) {
	done := make(chan struct{})
	rr := regreq{
		client: c,
		done:   done,
	}
	m.unregister <- rr
	<-done
}

// Run runs in its own goroutine processing (un)registration requests.
func (m *manager) Run(ctx context.Context) {
	// helper fn for cleaning up client
	cleanupClient := func(c Client) {
		clientCtx, ok := m.clients[c.ID()]
		if ok {
			clientCtx.Cancel()
			delete(m.clients, c.ID())
			clientCtx.Client().Close()
		}
	}

	for {
		select {
		case <-ctx.Done():
			m.mu.Lock()
			for _, client := range m.clients {
				cleanupClient(client.Client())
			}
			m.mu.Unlock()
		case rr := <-m.register:
			m.mu.Lock()
			m.clients[rr.client.ID()] = &clientContext{
				id:     rr.client.ID(),
				client: rr.client,
				cancel: rr.cancel,
			}
			m.mu.Unlock()
			rr.done <- struct{}{}

		case rr := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[rr.client.ID()]; ok {
				cleanupClient(rr.client)
			}
			m.mu.Unlock()
			rr.done <- struct{}{}
		}
	}
}
