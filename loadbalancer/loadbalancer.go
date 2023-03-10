// Package loadbalancer provides the internet-facing load balancer that is used
// to interact with applications.
package loadbalancer

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/maxmcd/steady/daemon"
	"github.com/maxmcd/steady/daemon/daemonrpc"
	"github.com/maxmcd/steady/internal/httpx"
	_ "github.com/maxmcd/steady/internal/slogx"
	"github.com/maxmcd/steady/internal/steadyutil"
	"github.com/maxmcd/steady/slicer"
	"github.com/pkg/errors"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

var (
	MaxHostAssignments = 10
)

type LB struct {
	eg              *errgroup.Group
	publicListener  net.Listener
	privateListener net.Listener
	client          *http.Client

	hashRangesSetLock *sync.Mutex
	hashRangesSet     []*slicer.HostAssignments
}

type Option func(*LB)

func NewLB(opts ...Option) *LB {
	lb := &LB{
		hashRangesSetLock: &sync.Mutex{},
		client:            &http.Client{},
	}
	for _, opt := range opts {
		opt(lb)
	}
	return lb
}

func (lb *LB) NewHostAssignments(assignments map[string][]slicer.Range) error {
	ha, err := slicer.NewHostAssignments(assignments)
	if err != nil {
		return err
	}
	lb.hashRangesSetLock.Lock()
	lb.hashRangesSet = append(lb.hashRangesSet, ha)
	// cap the number of things in the list
	if len(lb.hashRangesSet) > MaxHostAssignments {
		// Remove oldest generation
		lb.hashRangesSet = lb.hashRangesSet[1:]
	}
	lb.hashRangesSetLock.Unlock()
	return nil
}

func (lb *LB) Handler(name string, useLatestHost bool, rw http.ResponseWriter, r *http.Request) {
	var hosts []string
	for _, hashRanges := range lb.hashRangesSet {
		host := hashRanges.GetHost(name)
		// Add if it's not the previous host
		if len(hosts) == 0 || hosts[len(hosts)-1] != host {
			hosts = append(hosts, host)
		}
	}
	host := hosts[0]
	var err error
	if !useLatestHost && len(hosts) > 1 {
		host, err = lb.findLiveHost(r.Context(), hosts, name)
		if err != nil {
			http.Error(rw, errors.Wrap(err, "error routing to host").Error(),
				http.StatusBadGateway)
			return
		}
	}
	r.URL.Host = host
	r.URL.Scheme = "http"

	// Request.RequestURI can't be set in client requests.
	r.RequestURI = ""
	resp, err := lb.client.Do(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	for k, v := range resp.Header {
		rw.Header()[k] = v
	}
	rw.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(rw, resp.Body)
	_ = resp.Body.Close()
}

func (lb *LB) findLiveHost(ctx context.Context, hosts []string, name string) (host string, err error) {
	for _, host := range hosts {
		daemonClient := daemon.NewClient(fmt.Sprintf("http://%s", host), lb.client)
		if _, err = daemonClient.GetApplication(ctx, &daemonrpc.GetApplicationRequest{
			Name: name,
		}); err != nil {
			continue
		}
		// Success
		return host, nil
	}
	return "", err
}

// Wait until the server has stopped, returning any errors.
func (lb *LB) Wait() error {
	defer slog.Info("Load balancer stopped")
	return lb.eg.Wait()
}

// Start the server and listen at the provided address.
func (lb *LB) Start(ctx context.Context, publicAddr, privateAddr string) (err error) {
	if lb.eg != nil {
		return fmt.Errorf("LB has already started")
	}
	if len(lb.hashRangesSet) == 0 {
		return fmt.Errorf("can't start without host assignments")
	}

	lb.publicListener, err = net.Listen("tcp", publicAddr)
	if err != nil {
		return err
	}
	lb.privateListener, err = net.Listen("tcp", privateAddr)
	if err != nil {
		return err
	}
	slog.Info("load balancer listening",
		"private_addr", lb.privateListener.Addr().String(),
		"public_addr", lb.publicListener.Addr().String(),
	)

	lb.eg, ctx = errgroup.WithContext(ctx)

	publicServer := &http.Server{
		Handler: steadyutil.Logger("lb", os.Stdout,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				name := steadyutil.ExtractAppName(r)
				if name == "steady" {
					http.Error(w, "not allowed", http.StatusMethodNotAllowed)
					return
				}
				lb.Handler(name, false, w, r)
			}),
		),
		ReadHeaderTimeout: time.Second * 15,
	}
	lb.eg.Go(httpx.ServeContext(ctx, lb.publicListener, publicServer))

	privateServer := &http.Server{
		Handler: steadyutil.Logger("pb", os.Stdout, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: auth
			name := r.Header.Get(steadyutil.XAppName)
			if name == "" {
				http.Error(w, "not found", http.StatusNotFound)
			}
			lb.Handler(name, true, w, r)
		})),
		ReadHeaderTimeout: time.Second * 15,
	}
	lb.eg.Go(httpx.ServeContext(ctx, lb.privateListener, privateServer))

	return nil
}

// PublicServerAddr returns the address of the running server. Will panic if the
// server hasn't been started yet.
func (lb *LB) PublicServerAddr() string {
	if lb.eg == nil {
		panic(fmt.Errorf("server has not started"))
	}
	return strings.Replace(
		// For local dev, give a hostname so we can use subdomains
		lb.publicListener.Addr().String(),
		"127.0.0.1", "localhost", 1,
	)
}

// PrivateServerAddr returns the address of the running server. Will panic if
// the server hasn't been started yet.
func (lb *LB) PrivateServerAddr() string {
	if lb.eg == nil {
		panic(fmt.Errorf("server has not started"))
	}
	return lb.privateListener.Addr().String()
}
