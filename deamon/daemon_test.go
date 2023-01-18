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

func TestConcurrentRequests(t *testing.T) {
	w := NewWorkflow()
	timestamp := time.Now().Format(time.RFC3339)
	filename, err := w.writeApplicationScript(
		fmt.Sprintf(exampleServer, timestamp),
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error { return w.RunWorkerActivity(ctx, WorkerData{Filename: filename}) })

	requestCount := 5
	for i := 0; i < requestCount; i++ {
		j := i
		eg.Go(func() error {
			resp, err := http.Get("http://localhost:8080")
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

	assert.Equal(t, requestCount, w.workerState.requestCount)
	assert.Equal(t, 1, w.workerState.startCount)
}

func TestNonOverlappingTests(t *testing.T) {
	w := NewWorkflow()
	timestamp := time.Now().Format(time.RFC3339)
	filename, err := w.writeApplicationScript(
		fmt.Sprintf(exampleServer, timestamp),
	)
	if err != nil {
		t.Fatal(err)
	}

	makeRequest := func() {
		resp, err := http.Get("http://localhost:8080")
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)
		assert.Contains(t, buf.String(), timestamp)
	}

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error { return w.RunWorkerActivity(ctx, WorkerData{Filename: filename}) })

	makeRequest()

	// Wait for shutdown
	// TODO: parameterize
	time.Sleep(time.Second)

	makeRequest()
	cancel()

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, w.workerState.requestCount)
	assert.Equal(t, 2, w.workerState.startCount)
}

func BenchmarkActivity(b *testing.B) {
	w := NewWorkflow()
	w.requestLogger = io.Discard

	timestamp := time.Now().Format(time.RFC3339)
	filename, err := w.writeApplicationScript(
		fmt.Sprintf(exampleServer, timestamp),
	)
	if err != nil {
		b.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error { return w.RunWorkerActivity(ctx, WorkerData{Filename: filename}) })
	for i := 0; i < b.N; i++ {
		resp, err := http.Get("http://localhost:8080")
		if err != nil {
			b.Fatal(err)
		}
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)
		assert.Contains(b, buf.String(), timestamp)
	}
	cancel()
	if err := eg.Wait(); err != nil {
		b.Fatal(err)
	}
}
