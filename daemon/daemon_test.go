package daemon_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/maxmcd/steady/daemon"
	"github.com/maxmcd/steady/daemon/daemonrpc"
	"github.com/maxmcd/steady/internal/testsuite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

var exampleServer = `
let port = process.env.PORT ?? 3000;
console.log("Listening on port "+ port);
export default {
	port: port,
	fetch(request: Request): Response {
		return new Response("Hello %s" + request.url);
	},
};
`

type DaemonSuite struct {
	testsuite.Suite
}

func TestDaemonSuite(t *testing.T) {
	suite.Run(t, new(DaemonSuite))
}

func (suite *DaemonSuite) TestConcurrentRequests() {
	t := suite.T()
	d, _ := suite.NewDaemon()

	client := suite.NewDaemonClient(d.ServerAddr())
	timestamp := time.Now().Format(time.RFC3339)

	name := "max-hello"
	if _, err := client.CreateApplication(context.Background(), &daemonrpc.CreateApplicationRequest{
		Name:   name,
		Script: fmt.Sprintf(exampleServer, timestamp),
	}); err != nil {
		t.Fatal(err)
	}

	eg, _ := errgroup.WithContext(context.Background())

	requestCount := 5
	for i := 0; i < requestCount; i++ {
		eg.Go(func() error {
			resp, respBody, err := suite.DaemonRequest(d, name, http.MethodGet, "/hi", "")
			if err != nil {
				return err
			}
			suite.Equal(http.StatusOK, resp.StatusCode)
			suite.Contains(respBody, timestamp)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}

	app, err := client.GetApplication(context.Background(), &daemonrpc.GetApplicationRequest{
		Name: name,
	})
	if err != nil {
		t.Fatal(err)
	}

	suite.Equal(int64(requestCount), app.RequestCount)
	suite.Equal(int64(1), app.StartCount)
}

func (suite *DaemonSuite) TestNonOverlappingTests() {
	t := suite.T()
	d, _ := suite.NewDaemon()
	client := suite.NewDaemonClient(d.ServerAddr())
	timestamp := time.Now().Format(time.RFC3339)

	name := "max-hello"
	if _, err := client.CreateApplication(context.Background(), &daemonrpc.CreateApplicationRequest{
		Name:   name,
		Script: fmt.Sprintf(exampleServer, timestamp),
	}); err != nil {
		t.Fatal(err)
	}

	makeRequest := func() {
		resp, respBody, err := suite.DaemonRequest(d, name, http.MethodGet, "/hi", "")
		if err != nil {
			t.Fatal(err)
		}
		suite.Equal(http.StatusOK, resp.StatusCode)
		suite.Contains(respBody, timestamp)
	}

	makeRequest()

	d.StopAllApplications()

	makeRequest()

	app, err := client.GetApplication(context.Background(), &daemonrpc.GetApplicationRequest{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(2), app.RequestCount)
	assert.Equal(t, int64(2), app.StartCount)
}

func BenchmarkActivity(b *testing.B) {
	b.Skip("re-make this when we care")
	d := daemon.NewDaemon(b.TempDir(), "localhost:0")
	timestamp := time.Now().Format(time.RFC3339)

	client := daemon.NewClient(d.ServerAddr(), nil)
	name := "max-hello"
	if _, err := client.CreateApplication(context.Background(), &daemonrpc.CreateApplicationRequest{
		Name:   name,
		Script: fmt.Sprintf(exampleServer, timestamp),
	}); err != nil {
		b.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.Start(ctx)

	for i := 0; i < b.N; i++ {
		resp, err := http.Get("http://" + d.ServerAddr() + "/max-hello/hi")
		if err != nil {
			b.Fatal(err)
		}
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, resp.Body)
		assert.Contains(b, buf.String(), timestamp)
	}
	cancel()

	if err := d.Wait(); err != nil {
		b.Fatal(err)
	}
}
