package daemon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

var exampleServer = `
export default {
	port: process.env.PORT ?? 3000,
	fetch(request: Request): Response {
		return new Response("Hello %s " + request.url);
	},
};
`

func TestConcurrentRequests(t *testing.T) {
	d := NewDaemon(t.TempDir(), 8080)
	timestamp := time.Now().Format(time.RFC3339)

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)

	d.addApplication("max.hello", fmt.Sprintf(exampleServer, timestamp))

	d.Start(ctx)

	requestCount := 5
	for i := 0; i < requestCount; i++ {
		j := i
		eg.Go(func() error {
			resp, err := http.Get("http://localhost:8080/max.hello/hi")
			if err != nil {
				return err
			}
			var buf bytes.Buffer
			io.Copy(&buf, resp.Body)
			assert.Contains(t, buf.String(), timestamp)
			if j == 4 {
				cancel()
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := d.Wait(); err != nil {
		t.Fatal(err)
	}

	app := d.applications["max.hello"]
	assert.Equal(t, requestCount, app.requestCount)
	assert.Equal(t, 1, app.startCount)
}

func TestNonOverlappingTests(t *testing.T) {
	d := NewDaemon(t.TempDir(), 8080)
	timestamp := time.Now().Format(time.RFC3339)

	ctx, cancel := context.WithCancel(context.Background())
	d.addApplication("max.hello", fmt.Sprintf(exampleServer, timestamp))
	d.Start(ctx)

	makeRequest := func() {
		resp, err := http.Get("http://localhost:8080/max.hello/hi")
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)
		assert.Contains(t, buf.String(), timestamp)
	}

	makeRequest()

	// Wait for shutdown
	// TODO: parameterize delay time
	time.Sleep(time.Second)

	makeRequest()

	cancel()
	if err := d.Wait(); err != nil {
		t.Fatal(err)
	}

	app := d.applications["max.hello"]
	assert.Equal(t, 2, app.requestCount)
	assert.Equal(t, 2, app.startCount)
}

func BenchmarkActivity(b *testing.B) {
	d := NewDaemon(b.TempDir(), 8080)
	timestamp := time.Now().Format(time.RFC3339)

	d.addApplication("max.hello", fmt.Sprintf(exampleServer, timestamp))

	ctx, cancel := context.WithCancel(context.Background())
	d.Start(ctx)

	for i := 0; i < b.N; i++ {
		resp, err := http.Get("http://localhost:8080/max.hello/hi")
		if err != nil {
			b.Fatal(err)
		}
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)
		assert.Contains(b, buf.String(), timestamp)
	}
	cancel()

	if err := d.Wait(); err != nil {
		b.Fatal(err)
	}
}
