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
	"github.com/maxmcd/steady/internal/daemontest"
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
	daemontest.DaemonSuite
}

func TestDaemonSuite(t *testing.T) {
	suite.Run(t, new(DaemonSuite))
}

func (suite *DaemonSuite) TestConcurrentRequests() {
	t := suite.T()
	d, _ := suite.CreateDaemon()

	client := suite.NewClient(d)
	timestamp := time.Now().Format(time.RFC3339)

	name := "max.hello"
	_, err := client.CreateApplication(context.Background(), name, fmt.Sprintf(exampleServer, timestamp))
	if err != nil {
		t.Fatal(err)
	}

	eg, _ := errgroup.WithContext(context.Background())

	requestCount := 5
	for i := 0; i < requestCount; i++ {
		eg.Go(func() error {
			resp, err := http.Get(suite.DaemonURL(d, name, "hi"))
			if err != nil {
				return err
			}
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, resp.Body)
			suite.Contains(buf.String(), timestamp)
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("Unexpected HTTP status %d", resp.StatusCode)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}

	app, err := client.GetApplication(context.Background(), name)
	if err != nil {
		t.Fatal(err)
	}

	suite.Equal(requestCount, app.RequestCount)
	suite.Equal(1, app.StartCount)
}

func (suite *DaemonSuite) TestNonOverlappingTests() {

	t := suite.T()
	d, _ := suite.CreateDaemon()
	client := suite.NewClient(d)
	timestamp := time.Now().Format(time.RFC3339)

	name := "max.hello"
	if _, err := client.CreateApplication(context.Background(),
		name, fmt.Sprintf(exampleServer, timestamp)); err != nil {
		t.Fatal(err)
	}

	makeRequest := func() {
		resp, err := http.Get(suite.DaemonURL(d, name, "hi"))
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, resp.Body)
		suite.Equal(http.StatusOK, resp.StatusCode)
		assert.Contains(t, buf.String(), timestamp)
	}

	makeRequest()

	d.StopAllApplications()

	makeRequest()

	app, err := client.GetApplication(context.Background(), name)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, app.RequestCount)
	assert.Equal(t, 2, app.StartCount)
}

func BenchmarkActivity(b *testing.B) {
	d := daemon.NewDaemon(b.TempDir(), "localhost:0")
	timestamp := time.Now().Format(time.RFC3339)

	client, err := daemon.NewClient(d.ServerAddr(), nil)
	if err != nil {
		b.Fatal(err)
	}
	name := "max.hello"
	if _, err := client.CreateApplication(context.Background(), name, fmt.Sprintf(exampleServer, timestamp)); err != nil {
		b.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.Start(ctx)

	for i := 0; i < b.N; i++ {
		resp, err := http.Get("http://" + d.ServerAddr() + "/max.hello/hi")
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
