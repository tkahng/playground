package sse

import (
	"context"
	"errors"
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

type manager struct {
	mu         *sync.RWMutex
	clients    map[Client]context.CancelFunc
	register   chan regreq
	unregister chan regreq
}

// Send implements Manager.
func (m *manager) Send(clientId string, data any) error {

	for c := range m.clients {
		if c.ID() == clientId {
			return c.Write(Message{Data: data})
		}
	}
	return nil
}

// SendAll implements Manager.
func (m *manager) SendAll(data any) error {
	var errs []error
	for c := range m.clients {
		if err := c.Write(Message{Data: data}); err != nil {
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

func NewManager() Manager {
	return &manager{
		mu:         &sync.RWMutex{},
		clients:    make(map[Client]context.CancelFunc),
		register:   make(chan regreq),
		unregister: make(chan regreq),
	}
}

// Clients returns the currently managed Clients.
func (m *manager) Clients() []Client {
	res := []Client{}
	m.mu.RLock()
	defer m.mu.RUnlock()

	for c := range m.clients {
		res = append(res, c)
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
		cancel, ok := m.clients[c]
		if ok {
			cancel()
		}
		delete(m.clients, c)
		c.Close()
	}

	for {
		select {
		case <-ctx.Done():
			m.mu.Lock()
			for client := range m.clients {
				cleanupClient(client)
			}
			m.mu.Unlock()
		case rr := <-m.register:
			m.mu.Lock()
			m.clients[rr.client] = rr.cancel
			m.mu.Unlock()
			rr.done <- struct{}{}

		case rr := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[rr.client]; ok {
				cleanupClient(rr.client)
			}
			m.mu.Unlock()
			rr.done <- struct{}{}
		}
	}
}
