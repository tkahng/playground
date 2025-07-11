package sse

import (
	"fmt"
	"sync"
	"time"
)

// Producer is a struct that will produce messages to send to clients. It runs
// in a goroutine and produces the FizzBuzz sequence from 1 to 1000 over and
// over, generating one message per 500ms.
type Producer struct {
	EmitFunc func() any
	// Cancel is a channel that can be used to stop the producer.
	Cancel chan bool

	// clients keeps track of all connected clients, which enables us to send
	// messages to each of them when they are produced.
	clients []chan any

	// mu is a mutex to protect the clients slice from concurrent access.
	mu sync.Mutex
}

// AddClient adds a new client to the producer. Each message that is produced
// will be sent to each registered client.
func (p *Producer) AddClient(client chan any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.clients = append(p.clients, client)
}

// Produce the FizzBuzz sequence forever, starting from 1 again once 1000 is
// reached.
func (p *Producer) Produce() {
	for i := 1; true; i = (i + 1) % 1000 {
		// Emit our message!
		p.emit(p.EmitFunc())

		// Now, we want to either wait for 500ms or until we are canceled. If
		// canceled, we need to shut everything down nicely.
		select {
		case <-time.After(500 * time.Millisecond):
			// Not canceled, so continue the loop.
		case <-p.Cancel:
			fmt.Println("Stopping producer...")
			p.mu.Lock()
			defer p.mu.Unlock()
			for _, client := range p.clients {
				// Close each client channel to signal that we are done. This should
				// let each connected client disconnect gracefully.
				close(client)
			}
			return
		}
	}
}

// emit a message to all registered clients.
func (p *Producer) emit(data any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Loop backwards so we can remove dead clients without messing up the
	// indexes / traversal. This has a side-effect which may be undesirable of
	// the most recently connected clients getting the messages first.
	for i := len(p.clients) - 1; i >= 0; i-- {
		select {
		case p.clients[i] <- data:
			// Do nothing, send was successful.
		default:
			// Could not send...remove this client as it is dead!
			close(p.clients[i])
			p.clients = append(p.clients[:i], p.clients[i+1:]...)
		}
	}
}
