package daemon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

var exampleServer = `
export default {
	port: process.env.PORT ?? 3000,
	fetch(request: Request): Response {
		return new Response("Hello %s" + request.url);
	},
};
`

func (suite *DaemonSuite) TestConcurrentRequests() {
	timestamp := time.Now().Format(time.RFC3339)
	app, err := suite.d.validateAndAddApplication(
		"max.hello", []byte(fmt.Sprintf(exampleServer, timestamp)))
	suite.NoError(err)

	eg, _ := errgroup.WithContext(context.Background())

	fmt.Println(app.port)

	requestCount := 5
	for i := 0; i < requestCount; i++ {
		eg.Go(func() error {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/max.hello/hi", suite.port))
			if err != nil {
				return err
			}
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, resp.Body)
			suite.Contains(buf.String(), timestamp)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		suite.T().Fatal(err)
	}

	suite.Equal(requestCount, app.requestCount)
	suite.Equal(1, app.startCount)
}

func TestNonOverlappingTests(t *testing.T) {
	d := NewDaemon(t.TempDir(), 8080)
	timestamp := time.Now().Format(time.RFC3339)

	ctx, cancel := context.WithCancel(context.Background())
	if _, err := d.validateAndAddApplication("max.hello", []byte(fmt.Sprintf(exampleServer, timestamp))); err != nil {
		t.Fatal(err)
	}
	d.Start(ctx)

	makeRequest := func() {
		resp, err := http.Get("http://localhost:8080/max.hello/hi")
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, resp.Body)
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

	if _, err := d.validateAndAddApplication("max.hello", []byte(fmt.Sprintf(exampleServer, timestamp))); err != nil {
		b.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.Start(ctx)

	for i := 0; i < b.N; i++ {
		resp, err := http.Get("http://localhost:8080/max.hello/hi")
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

func (suite *DaemonSuite) TestCreateApplication() {
	client := suite.newClient()

	app, err := client.CreateApplication("max.db", exampleServer)
	suite.Require().NoError(err)

	fmt.Println(app.Name)
}

func Test_bunRun(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr bool
	}{
		{"junk script", "asdfasdf", true},
		{"no server", "console.log('hi')", true},
		{"wrong port", `export default { port: 12345, fetch(request) { return new Response("Hello")} };`, true},
		{"a good one", `export default { port: process.env.PORT, fetch(request) { return new Response("Hello")} };`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			f, err := os.Create(filepath.Join(dir, "index.ts"))
			if err != nil {
				t.Fatal(err)
			}
			_, _ = f.Write([]byte(tt.script))
			_ = f.Close()
			port, err := getFreePort()
			if err != nil {
				t.Fatal(err)
			}
			got, err := bunRun(dir, port, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("bunRun() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && got.Process != nil {
				if err := got.Process.Kill(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
