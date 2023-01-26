package loadbalancer

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/maxmcd/steady/daemon"
	"github.com/maxmcd/steady/slicer"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var (
	maxAssignments = 10
)

type LB struct {
	appNameExtractor AppNameExtractor

	eg           *errgroup.Group
	listenerWait *sync.WaitGroup
	listener     net.Listener
	client       *http.Client

	hashRangesSetLock *sync.Mutex
	hashRangesSet     []*slicer.HostAssignments
}

type Option func(*LB)

func OptionWithAppNameExtractor(a AppNameExtractor) func(*LB) {
	return func(l *LB) { l.appNameExtractor = a }
}

func NewLB(opts ...Option) *LB {
	lb := &LB{
		listenerWait:      &sync.WaitGroup{},
		hashRangesSetLock: &sync.Mutex{},
		client:            &http.Client{},
	}
	lb.listenerWait.Add(1)
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
	if len(lb.hashRangesSet) > maxAssignments {
		// Remove oldest generation
		lb.hashRangesSet = lb.hashRangesSet[1:]
	}
	lb.hashRangesSetLock.Unlock()
	return nil
}

func (lb *LB) Handler(rw http.ResponseWriter, r *http.Request) {
	name, err := lb.appNameExtractor(r)
	if err != nil {
		http.Error(rw, errors.Wrap(err, "error finding application name").Error(),
			http.StatusBadGateway)
		return
	}
	var hosts []string
	for _, hashRanges := range lb.hashRangesSet {
		host := hashRanges.GetHost(name).Host
		// Add if it's not the previous host
		if len(hosts) == 0 || hosts[len(hosts)-1] != host {
			hosts = append(hosts, host)
		}
	}
	host := hosts[0]
	if len(hosts) > 1 {
		host, err = lb.findLiveHost(r.Context(), hosts, name)
		if err != nil {
			http.Error(rw, errors.Wrap(err, "error routing to host").Error(),
				http.StatusBadGateway)
			return
		}
	}
	fmt.Println("HAWST", host)

	r.URL.Host = host
	r.URL.Scheme = "http"
	r.URL.Path = "/" + name + r.URL.Path
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
		fmt.Println(host, hosts)
		var daemonClient *daemon.Client
		if daemonClient, err = daemon.NewClient(fmt.Sprintf("http://%s", host), lb.client); err != nil {
			// Shouldn't error here ever
			return "", err
		}

		if _, err = daemonClient.GetApplication(ctx, name); err != nil {
			continue
		}
		// Success
		return host, nil
	}
	return "", err
}

// Wait until the server has stopped, returning any errors.
func (lb *LB) Wait() error { return lb.eg.Wait() }

// Start the server and listen at the provided address.
func (lb *LB) Start(ctx context.Context, addr string) {
	if lb.eg != nil {
		panic("LB has already started")
	}
	if len(lb.hashRangesSet) == 0 {
		panic("can't start without host assignments")
	}

	lb.eg, ctx = errgroup.WithContext(ctx)

	srv := http.Server{
		Handler:           http.HandlerFunc(lb.Handler),
		ReadHeaderTimeout: time.Second * 15,
	}

	lb.eg.Go(func() (err error) {
		lb.listener, err = net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		lb.listenerWait.Done()
		err = srv.Serve(lb.listener)
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	})
	lb.eg.Go(func() error {
		<-ctx.Done()
		timeoutCtx, _ := context.WithTimeout(ctx, time.Minute) // TODO: vet this time
		if err := srv.Shutdown(timeoutCtx); err != nil {
			fmt.Printf("WARN: error shutting down http server: %v\n", err)
		}
		return nil
	})
}

func (lb *LB) ServerAddr() string {
	if lb.eg == nil {
		panic(fmt.Errorf("server has not started"))
	}
	lb.listenerWait.Wait()
	return lb.listener.Addr().String()
}

type AppNameExtractor func(req *http.Request) (string, error)

var _ AppNameExtractor = TestHeaderExtractor

func TestHeaderExtractor(req *http.Request) (string, error) {
	host := req.Header.Get("X-Host")
	if host == "" {
		return host, errors.New("app name not found")
	}
	return host, nil
}
