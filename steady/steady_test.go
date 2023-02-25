package steady_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	daemonrpc "github.com/maxmcd/steady/daemon/daemonrpc"
	"github.com/maxmcd/steady/internal/testsuite"
	"github.com/maxmcd/steady/loadbalancer"
	"github.com/sourcegraph/conc/pool"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/slog"
)

type TestSuite struct {
	testsuite.Suite
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) TestMoveApplication() {
	t := suite.T()
	// Migrate job
	// Start job on daemon
	// send requests to it from the load balancer
	// add another host
	// migrate the job to another daemon
	// ensure all requests make it to a live job
	appName := "whee"
	httpClient := &http.Client{}

	suite.StartMinioServer()
	d, _ := suite.NewDaemon()
	lb := suite.NewLB()
	ctx := context.Background()
	{
		// Confirm the application currently returns a 404
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s", lb.PublicServerAddr()), nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Host", appName)
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		suite.Equal(http.StatusNotFound, resp.StatusCode)
	}

	dClient := suite.NewDaemonClient(d.ServerAddr())
	if _, err := dClient.CreateApplication(ctx, &daemonrpc.CreateApplicationRequest{
		Name:   appName,
		Script: suite.LoadExampleScript("http"),
	}); err != nil {
		t.Fatal(err)
	}

	counter := 0
	createRecordRequest := func() {
		req, err := http.NewRequest(http.MethodPost,
			fmt.Sprintf("http://%s", lb.PublicServerAddr()),
			bytes.NewBuffer([]byte(`{"email":"lite"}`)))
		if err != nil {
			t.Fatal(err)
		}
		req.Host = appName
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var jsonResponse map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&jsonResponse)
		_ = resp.Body.Close()

		suite.Require().Equal(http.StatusOK, resp.StatusCode, jsonResponse)

		// Here's the real test. Ensure that every time we make a request the ID
		// increments, even through deletes/restarts
		counter++
		suite.Require().Equal(counter, int(jsonResponse["id"].(float64)), jsonResponse)
	}
	createRecordRequest()
	d.StopAllApplications()

	d2, _ := suite.NewDaemon()
	d2Client := suite.NewDaemonClient(d2.ServerAddr())

	// How are we finding what to move?
	if _, err := dClient.DeleteApplication(ctx, &daemonrpc.DeleteApplicationRequest{Name: appName}); err != nil {
		t.Fatal(err)
	}

	if _, err := d2Client.CreateApplication(ctx, &daemonrpc.CreateApplicationRequest{
		Name:   appName,
		Script: suite.LoadExampleScript("http"),
	}); err != nil {
		t.Fatal(err)
	}
	createRecordRequest()
}

func (suite *TestSuite) TestUpdateApplication() {
	t := suite.T()
	suite.StartMinioServer()
	d, _ := suite.NewDaemon()
	lb := suite.NewLB()
	ctx := context.Background()

	appName := "foo"

	dClient := suite.NewDaemonClient(d.ServerAddr())
	if _, err := dClient.CreateApplication(ctx, &daemonrpc.CreateApplicationRequest{
		Name:   appName,
		Script: suite.LoadExampleScript("http"),
	}); err != nil {
		t.Fatal(err)
	}
	rm := &RequestMaker{lb: lb, appName: appName}
	stop := rm.runMany(1)
	rm.Request(ctx)
	rm.Request(ctx)
	rm.Request(ctx)

	_, err := dClient.UpdateApplication(ctx, &daemonrpc.UpdateApplicationRequest{
		Name:   appName,
		Script: suite.LoadExampleScript("http"),
	})
	if err != nil {
		t.Fatal(err)
	}
	rm.Request(ctx)
	rm.Request(ctx)
	rm.Request(ctx)
	stop()

	suite.Equal(atomic.LoadInt32(&rm.counterValue), atomic.LoadInt32(&rm.successfulRequestCount))
	fmt.Println(atomic.LoadInt32(&rm.requestCount), atomic.LoadInt32(&rm.successfulRequestCount))
}

type RequestMaker struct {
	counterValue           int32
	successfulRequestCount int32
	requestCount           int32

	appName string
	lb      *loadbalancer.LB
}

func (rm *RequestMaker) runMany(threads int) func() {
	ctx, cancel := context.WithCancel(context.Background())
	p := pool.New().WithMaxGoroutines(threads).WithContext(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			if ctx.Err() != nil {
				wg.Done()
				return
			}
			time.Sleep(time.Millisecond * 1)
			p.Go(func(ctx context.Context) error {
				rm.Request(ctx)
				return nil
			})
		}
	}()
	return func() {
		// Use these values so that cancelled requests are not included
		a, b, c := rm.counterValue, rm.successfulRequestCount, rm.requestCount
		cancel()
		wg.Wait()
		atomic.StoreInt32(&rm.counterValue, a)
		atomic.StoreInt32(&rm.successfulRequestCount, b)
		atomic.StoreInt32(&rm.requestCount, c)
	}
}

func (rm *RequestMaker) Request(ctx context.Context) {
	atomic.AddInt32(&rm.requestCount, 1)

	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("http://%s", rm.lb.PublicServerAddr()),
		bytes.NewBuffer([]byte(`{"email":"lite"}`)))
	if err != nil {
		slog.Error("request error", err)
		return
	}
	req = req.WithContext(ctx)
	req.Host = rm.appName
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("request error", err)
		return
	}

	var jsonResponse map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	_ = resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		atomic.AddInt32(&rm.successfulRequestCount, 1)
	} else {
		slog.Error("incorrect status code", nil, "code", resp.StatusCode)
	}
	atomic.StoreInt32(&rm.counterValue, int32(jsonResponse["id"].(float64)))
}
