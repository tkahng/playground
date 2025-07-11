package sse_test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	humasse "github.com/danielgtaylor/huma/v2/sse"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/tools/sse"
)

type DefaultMessage struct {
	Message string `json:"message"`
}

func TestWSHandler(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	_, api := humatest.New(t)

	// synchronization helpers
	doneReg := make(chan sse.Client, 1)
	doneUnreg := make(chan sse.Client, 1)
	messageChan := make(chan struct{}, 1)
	var c sse.Client
	var cf context.CancelFunc

	manager := sse.NewManager(slog.Default())
	go manager.Run(ctx)

	h := sse.ServeSSE[struct{}](
		func(ctx context.Context, f func(any) error) sse.Client {
			return sse.NewClient("test", func(a any) error {
				err := f(a)
				messageChan <- struct{}{}
				return err
			}, slog.Default())
		},
		func(ctx context.Context, _cf context.CancelFunc, _c sse.Client) {
			cf = _cf
			c = _c
			manager.RegisterClient(ctx, cf, c)
			doneReg <- c
		},
		func(_c sse.Client) {
			t.Log("unregistering client in ondestroy")
			// manager.UnregisterClient(_c)
			t.Log("unregistered client in ondestroy")
			t.Log("waiting for client to close in ondestroy")
			_c.Wait()
			t.Log("client closed in ondestroy")
			doneUnreg <- _c
		},
		50*time.Second,
	)

	// setup and connect to the the test server using a basic websocket
	humasse.Register(
		api,
		huma.Operation{
			OperationID: "sse",
			Method:      http.MethodGet,
			Path:        "/sse",
			Summary:     "Server sent events example",
		},
		map[string]any{
			"message": &DefaultMessage{},
		},
		h,
	)

	// once registration is done, the manager should have one client
	t.Log("waiting for registration")
	resp := api.Get("/sse")
	<-doneReg
	t.Log("registration done")
	assert.Equal(t, len(manager.Clients()), 1)

	err := manager.Send("test", DefaultMessage{Message: "test"})
	// err := c.Write(sse.Message{
	// 	Data: DefaultMessage{Message: "Hello, world!"},
	// })
	<-messageChan
	assert.NoError(t, err)
	// write a message to the server; this will be echoed back
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "text/event-stream", resp.Header().Get("Content-Type"))
	assert.Equal(t, `data: {"message":"test"}

`, resp.Body.String())

	cancel()
	_p := <-doneUnreg
	assert.Equal(t, len(manager.Clients()), 0)
	assert.Equal(t, _p, c)
	// time.Sleep(1 * time.Second)
	//FIXME: seems to be leaking goroutines
}
