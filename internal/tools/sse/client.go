package sse

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

type Message struct {
	Data any `json:"data"`
}

type Client interface {

	// WriteForever is responsible for writing messages to the client (including
	// the regularly spaced ping messages)
	WriteForever(context.Context, func(Client), time.Duration)
	// Wait blocks until the client is done processing messages
	Wait()

	// Client Key is the unique client identifier. usually some kind of keyName:keyValue
	Channel() string

	// write is a low level function to send messages to the client
	Write(Message) error

	// Close implements the Closer interface. Note the behavior of calling Close()
	// multiple times is undefined; this implementation swallows all errors.
	Close() error
}
type client struct {
	lock             *sync.RWMutex
	wg               *sync.WaitGroup
	egress           chan Message
	logger           *slog.Logger
	channel          string
	send             func(any) error
	pingMessageFunc  func() any
	closeMessageFunc func() any
}

func NewClient(clientId string, sender func(any) error, logger *slog.Logger, pingMessageFunc func() any) Client {
	// add 2 to the wait group for the read/write goroutines
	wg := &sync.WaitGroup{}
	wg.Add(1)
	return &client{
		lock:            &sync.RWMutex{},
		wg:              wg,
		send:            sender,
		channel:         clientId,
		egress:          make(chan Message, 32),
		logger:          logger,
		pingMessageFunc: pingMessageFunc,
	}
}

func (c *client) Channel() string {
	return c.channel
}

// Write implements the Writer interface.
func (c *client) Write(p Message) error {
	c.egress <- p
	return nil
}

// Close implements the Closer interface. Note the behavior of calling Close()
// multiple times is undefined; this implementation swallows all errors.
func (c *client) Close() error {

	return nil
}

// WriteForever serially processes messages from the egress channel and writes them
// to the client, ensuring that all writes to the underlying connection are
// performed here.
func (c *client) WriteForever(ctx context.Context, onDestroy func(Client), ping time.Duration) {
	if c.send == nil {
		c.Log(int(slog.LevelError), "no send function provided")
	}

	pingTicker := time.NewTicker(ping)
	defer func() {
		c.wg.Done()
		pingTicker.Stop()
		onDestroy(c)
	}()
	for {
		select {
		case <-ctx.Done():
			if c.closeMessageFunc != nil {
				if err := c.send(c.closeMessageFunc()); err != nil {
					c.Log(int(slog.LevelError), fmt.Sprintf("error writing close message: %v", err))
				}
			}
			return
		case message, ok := <-c.egress:
			// ok will be false in case the egress channel is closed
			if !ok {
				if c.closeMessageFunc != nil {
					if err := c.send(c.closeMessageFunc()); err != nil {
						c.Log(int(slog.LevelError), fmt.Sprintf("error writing close message: %v", err))
					}
				}
				return
			}
			// write a message to the connection
			if err := c.send(message.Data); err != nil {
				c.Log(int(slog.LevelError), fmt.Sprintf("error writing message: %v", err))
				return
			}
		case <-pingTicker.C:
			if c.pingMessageFunc != nil {
				if err := c.send(c.pingMessageFunc()); err != nil {
					c.Log(int(slog.LevelError), fmt.Sprintf("error writing ping: %v", err))
					return
				}
			}
		}
	}
}

func (c *client) SetLogger(v any) error {
	l, ok := v.(*slog.Logger)
	if !ok {
		return fmt.Errorf("bad logger value supplied")
	}
	c.logger = l
	return nil
}

func (c *client) Log(level int, s string, args ...any) {
	_, f, l, ok := runtime.Caller(1)
	if ok {
		args = append(args, "caller_source", fmt.Sprintf("%s %d", f, l))
	}
	switch level {
	case int(slog.LevelDebug):
		c.logger.Debug(s, args...)
	case int(slog.LevelInfo):
		c.logger.Info(s, args...)
	case int(slog.LevelWarn):
		c.logger.Warn(s, args...)
	case int(slog.LevelError):
		c.logger.Error(s, args...)
	}
}

// Done blocks until the read/write goroutines have completed
func (c *client) Wait() {
	c.wg.Wait()
}
