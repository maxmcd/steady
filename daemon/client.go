package daemon

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/maxmcd/steady/daemon/daemonrpc"
	"github.com/maxmcd/steady/internal/steadyutil"
	"github.com/twitchtv/twirp"
)

// Client is an http client that is used to communicate with host daemons.
type Client struct {
	daemon daemonrpc.Daemon
}

func NewClient(server string, httpClient *http.Client) Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
			},
		}
	}
	return Client{daemon: daemonrpc.NewDaemonProtobufClient(server, httpClient, twirp.WithClientHooks(&twirp.ClientHooks{
		RequestPrepared: func(ctx context.Context, r *http.Request) (_ context.Context, err error) {
			r.Host = "steady"
			return ctx, nil
		},
	}))}
}

type getNamer interface {
	GetName() string
}

func addHeader(ctx context.Context, req getNamer) context.Context {
	header := make(http.Header)
	header.Set(steadyutil.XAppName, req.GetName())
	ctx, _ = twirp.WithHTTPRequestHeaders(ctx, header)
	// Ok to ignore this err, check function src
	return ctx
}
func (c Client) CreateApplication(ctx context.Context, req *daemonrpc.CreateApplicationRequest) (
	*daemonrpc.Application, error) {
	return c.daemon.CreateApplication(addHeader(ctx, req), req)
}

func (c Client) DeleteApplication(ctx context.Context, req *daemonrpc.DeleteApplicationRequest) (
	*daemonrpc.Application, error) {
	return c.daemon.DeleteApplication(addHeader(ctx, req), req)
}

func (c Client) GetApplication(ctx context.Context, req *daemonrpc.GetApplicationRequest) (
	*daemonrpc.Application, error) {
	return c.daemon.GetApplication(addHeader(ctx, req), req)
}
