//go:build !integration

package sse_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/tools/sse"
)

func TestManager_IsChannelConnected_FalseWhenEmpty(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := sse.NewManager(slog.Default())
	go m.Run(ctx)

	assert.False(t, m.IsChannelConnected("player_id:some-id"))
}

func TestManager_IsChannelConnected_TrueAfterRegister(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := sse.NewManager(slog.Default())
	go m.Run(ctx)

	clientCtx, clientCancel := context.WithCancel(ctx)
	client := sse.NewClient("player_id:abc", func(any) error { return nil }, slog.Default(), nil)
	m.RegisterClient(clientCtx, clientCancel, client)

	assert.True(t, m.IsChannelConnected("player_id:abc"))
	// Different channel must not match.
	assert.False(t, m.IsChannelConnected("player_id:xyz"))

	// UnregisterClient blocks until the manager has processed the removal,
	// so no sleep is needed before the subsequent assertion.
	clientCancel()
	m.UnregisterClient(client)

	assert.False(t, m.IsChannelConnected("player_id:abc"))
}

func TestManager_IsChannelConnected_FalseAfterUnregister(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := sse.NewManager(slog.Default())
	go m.Run(ctx)

	clientCtx, clientCancel := context.WithCancel(ctx)
	client := sse.NewClient("player_id:test", func(any) error { return nil }, slog.Default(), nil)
	m.RegisterClient(clientCtx, clientCancel, client)
	assert.True(t, m.IsChannelConnected("player_id:test"))

	clientCancel()
	m.UnregisterClient(client)

	assert.False(t, m.IsChannelConnected("player_id:test"))
}

func TestManager_IsChannelConnected_TrueWhileOneOfTwoClientsConnected(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := sse.NewManager(slog.Default())
	go m.Run(ctx)

	const channel = "player_id:multi"
	noop := func(any) error { return nil }

	ctx1, cancel1 := context.WithCancel(ctx)
	c1 := sse.NewClient(channel, noop, slog.Default(), nil)
	m.RegisterClient(ctx1, cancel1, c1)

	ctx2, cancel2 := context.WithCancel(ctx)
	c2 := sse.NewClient(channel, noop, slog.Default(), nil)
	m.RegisterClient(ctx2, cancel2, c2)

	assert.True(t, m.IsChannelConnected(channel))

	// Remove first client — channel still has c2.
	cancel1()
	m.UnregisterClient(c1)
	assert.True(t, m.IsChannelConnected(channel))

	// Remove second client — channel is now empty.
	cancel2()
	m.UnregisterClient(c2)
	assert.False(t, m.IsChannelConnected(channel))
}
