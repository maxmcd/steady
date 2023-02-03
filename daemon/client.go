package daemon

import (
	"net"
	"net/http"
	"time"

	"github.com/maxmcd/steady/daemon/rpc"
)

// Client is an http client that is used to communicate with host daemons.
type Client rpc.Daemon

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
	return rpc.NewDaemonProtobufClient(server, httpClient)
}
