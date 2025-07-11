package sse

import (
	"context"
	"time"

	humasse "github.com/danielgtaylor/huma/v2/sse"
)

func ServeSSE[I any](

	// clientFactory is a function that takes a connection and returns a new Client
	clientFactory func(context.Context, func(any) error) Client,
	// onCreate is a function to call once the Client is created (e.g.,
	// store it in a some collection on the service for later reference)
	onCreate func(context.Context, context.CancelFunc, Client),
	// onDestroy is a function to call after the WebSocket connection is closed
	// (e.g., remove it from the collection on the service)
	onDestroy func(Client),
	// ping is the interval at which ping messages are aren't sent
	ping time.Duration,
	// msgHandlers are callbacks that handle messages received from the client
) func(context.Context, *I, humasse.Sender) {
	return func(ctx context.Context, input *I, send humasse.Sender) {

		baseCtx, cf := context.WithCancel(context.Background())
		client := clientFactory(ctx, send.Data)
		onCreate(baseCtx, cf, client)
		// defer func() {
		// 	cf()
		// }()
		// all writes will happen in this goroutine, ensuring only one write on
		// the connection at a time
		client.WriteForever(baseCtx, onDestroy, ping)
	}
}
