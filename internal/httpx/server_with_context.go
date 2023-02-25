package httpx

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

func ServeContext(ctx context.Context, l net.Listener, server *http.Server) (wait func() error) {
	eg, ctx := errgroup.WithContext(ctx)
	lock := &sync.Mutex{}
	var closing bool
	eg.Go(func() (err error) {
		err = server.Serve(l)
		lock.Lock()
		defer lock.Unlock()
		if err == http.ErrServerClosed && closing {
			return nil
		}
		return err
	})
	eg.Go(func() error {
		<-ctx.Done()
		lock.Lock()
		closing = true
		lock.Unlock()
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Minute) // TODO: vet this time
		if err := server.Shutdown(timeoutCtx); err != nil {
			slog.Error("shutting down web server", err)
		}
		cancel()
		return nil
	})
	return eg.Wait
}
