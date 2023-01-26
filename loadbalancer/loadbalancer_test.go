package loadbalancer

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/maxmcd/steady/slicer"
	"github.com/stretchr/testify/require"
)

func TestLB(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	lb := NewLB(OptionWithAppNameExtractor(TestHeaderExtractor))

	shutdownApplication := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)
		if shutdownApplication {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	uri, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := lb.NewHostAssignments(map[string][]slicer.Range{
		uri.Host: {{0, math.MaxInt64}}},
	); err != nil {
		t.Fatal(err)
	}

	lb.Start(ctx, ":0")
	addr := lb.ServerAddr()

	makeRequest := func() {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s", addr), nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Host", "max.db")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(resp.Body)
		fmt.Println(string(b))
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}
	makeRequest()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))

	uri2, err := url.Parse(server2.URL)
	if err != nil {
		t.Fatal(err)
	}

	if err := lb.NewHostAssignments(map[string][]slicer.Range{
		uri2.Host: {{0, math.MaxInt64 / 2}},
		uri.Host:  {{(math.MaxInt64 / 2) + 1, math.MaxInt64}},
	}); err != nil {
		t.Fatal(err)
	}
	shutdownApplication = true
	t.Log("happening")
	makeRequest()

	cancel()
	if err := lb.Wait(); err != nil {
		t.Fatal(err)
	}

}
