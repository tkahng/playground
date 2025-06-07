package notifier

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/matryer/is"
	"github.com/tkahng/authgo/internal/test"
)

// NB: these tests assume you have a postgres server listening on localhost:5432
// with username postgres and password postgres. You can trivially set this up
// with Docker with the following:

// docker run --rm --name postgres -p 5432:5432 \
// -e POSTGRES_PASSWORD=postgres postgres

func TestNotifier(t *testing.T) {
	ctx, dbx := test.DbSetup()
	expIs := is.New(t)
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, cancel := context.WithCancel(ctx)
	wg := sync.WaitGroup{}
	// pool, err := testPool("postgres://postgres:postgres@localhost:5432/authgo_test?sslmode=disable")
	// expIs.NoErr(err)

	li := NewListener(dbx.Pool())
	err := li.Connect(ctx)
	expIs.NoErr(err)

	n := NewNotifier(l, li)
	wg.Add(1)
	go func() {
		err := n.Run(ctx)
		if err != nil {
			return
		}
		wg.Done()
	}()
	sub := n.Listen("foo")

	conn, err := dbx.Pool().Acquire(ctx)
	wg.Add(1)
	go func() {
		<-sub.EstablishedC()
		_, err := conn.Exec(ctx, "select pg_notify('foo', '1')")
		if err != nil {
			return
		}
		_, err = conn.Exec(ctx, "select pg_notify('foo', '2')")
		if err != nil {
			return
		}
		_, err = conn.Exec(ctx, "select pg_notify('foo', '3')")
		if err != nil {
			return
		}
		_, err = conn.Exec(ctx, "select pg_notify('foo', '4')")
		if err != nil {
			return
		}
		_, err = conn.Exec(ctx, "select pg_notify('foo', '5')")
		if err != nil {
			return
		}
		wg.Done()
	}()
	expIs.NoErr(err)

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
	expIs.Equal(msgs, []string{"1", "2", "3", "4", "5"})

	cancel()
	sub.Unlisten(ctx) // uses background ctx anyway
	err = li.Close(ctx)
	if err != nil {
		return
	}
	wg.Wait()
}
