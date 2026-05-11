package sse_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

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

	clientCancel()
	m.UnregisterClient(client)
	// Allow the manager's Run loop to process the unregister.
	time.Sleep(10 * time.Millisecond)

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
	time.Sleep(10 * time.Millisecond)

	assert.False(t, m.IsChannelConnected("player_id:test"))
}
