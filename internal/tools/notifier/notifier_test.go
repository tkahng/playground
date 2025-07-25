package notifier

import (
	"context"
	"log/slog"
	"os"
	"slices"
	"sync"
	"testing"

	"github.com/tkahng/playground/internal/test"
)

// NB: these tests assume you have a postgres server listening on localhost:5432
// with username postgres and password postgres. You can trivially set this up
// with Docker with the following:

// docker run --rm --name postgres -p 5432:5432 \
// -e POSTGRES_PASSWORD=postgres postgres

func TestNotifier(t *testing.T) {
	ctx, dbx := test.DbSetup()
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, cancel := context.WithCancel(ctx)
	wg := sync.WaitGroup{}
	// pool, err := testPool("postgres://postgres:postgres@localhost:5432/playground_test?sslmode=disable")
	// expIs.NoErr(err)

	li := NewListener(dbx)
	err := li.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	n := NewNotifier(l, li)
	wg.Add(1)
	go func() {
		n.Run(ctx)
		wg.Done()
	}()
	sub := n.Subscribe("foo")

	conn, err := dbx.Pool().Acquire(ctx)
	wg.Add(1)
	go func() {
		<-sub.EstablishedC()
		conn.Exec(ctx, "select pg_notify('foo', '1')")
		conn.Exec(ctx, "select pg_notify('foo', '2')")
		conn.Exec(ctx, "select pg_notify('foo', '3')")
		conn.Exec(ctx, "select pg_notify('foo', '4')")
		conn.Exec(ctx, "select pg_notify('foo', '5')")
		wg.Done()
	}()
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)

	out := make(chan string)
	go func() {
		<-sub.EstablishedC()
		for i := 0; i < 5; i++ {
			msg := <-sub.NotificationC()
			out <- string(msg)
		}
		close(out)
		wg.Done()
	}()

	var msgs []string
	for r := range out {
		msgs = append(msgs, r)
	}
	want := []string{"1", "2", "3", "4", "5"}
	if !slices.Equal(msgs, want) {
		t.Fatal("expected different order")
	}

	cancel()
	sub.Unlisten(ctx) // uses background ctx anyway
	li.Close(ctx)
	wg.Wait()
}
