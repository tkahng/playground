package sse

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type ClientContext interface {
	ID() string
	Client() Client
	Cancel() context.CancelFunc
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

type Manager interface {
	Clients() []Client
	RegisterClient(context.Context, context.CancelFunc, Client)
	UnregisterClient(Client)
	Run(context.Context)

	Send(clientId string, data any) error
	SendAll(data any) error
}

type manager struct {
	logger     *slog.Logger
	mu         *sync.RWMutex
	clients    map[Client]context.CancelFunc
	register   chan regreq
	unregister chan regreq
}

// Send implements Manager.
func (m *manager) Send(channel string, data any) error {
	var errs []error
	for c := range m.clients {
		if c.Channel() == channel {
			m.logger.Debug("client found", "channel", channel)
			err := c.Write(Message{Data: data})
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
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

func NewManager(logger *slog.Logger) Manager {
	return &manager{
		mu:         &sync.RWMutex{},
		logger:     logger,
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
	m.logger.Info("registering client")
	m.register <- rr
	m.logger.Info("client registered")
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
			m.logger.Info("shutting down")
			m.mu.Lock()
			for c := range m.clients {
				cleanupClient(c)
			}
			m.mu.Unlock()
			m.logger.Info("shutdown complete")
			return

		case rr := <-m.register:
			fmt.Println("got register")
			m.mu.Lock()
			m.clients[rr.client] = rr.cancel
			m.mu.Unlock()
			fmt.Println("register done")
			rr.done <- struct{}{}

		case rr := <-m.unregister:
			fmt.Println("got unregister")
			m.mu.Lock()
			if _, ok := m.clients[rr.client]; ok {
				cleanupClient(rr.client)
			}
			m.mu.Unlock()
			fmt.Println("unregister done")
			rr.done <- struct{}{}
		}
	}
}
