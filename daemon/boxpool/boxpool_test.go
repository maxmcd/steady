package boxpool_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/docker/docker/client"
	"github.com/maxmcd/steady/daemon/boxpool"
	"github.com/maxmcd/steady/internal/steadyutil"
	"github.com/maxmcd/steady/internal/testsuite"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	dataDir := t.TempDir()

	suite := new(testsuite.Suite)
	suite.SetT(t)
	pool, err := boxpool.New(context.Background(), "runner", dataDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Shutdown)

	appDir := t.TempDir()
	f, err := os.Create(filepath.Join(appDir, "index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(suite.LoadExampleScript("http"))
	f.Close()

	for i := 0; i < 5; i++ {
		fmt.Println("New run")
		healthEndpoint := steadyutil.RandomString(10)

		start := time.Now()
		box, err := pool.RunBox(context.Background(),
			[]string{"bun", "run", "/home/steady/wrapper.ts", "--no-install"},
			appDir,
			[]string{
				"STEADY_INDEX_LOCATION=/opt/app/index.ts",
				"STEADY_HEALTH_ENDPOINT=/" + healthEndpoint,
			})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("start", time.Since(start))
		start = time.Now()

		for i := 0; i < 20; i++ {
			res, err := http.Get(fmt.Sprintf("http://%s:3000/"+healthEndpoint, box.IPAddress()))
			if err == nil {
				_, _ = io.Copy(os.Stdout, res.Body)
				_ = res.Body.Close()
				break
			}
			exponent := time.Duration((i+1)*(i+1)) / 2
			time.Sleep(time.Millisecond * exponent)
			_, running, err := box.Status()
			if err != nil {
				t.Fatal(err)
			}
			if !running {
				t.Fatal("not running")
			}
		}

		fmt.Println("alive", time.Since(start))
		start = time.Now()

		if err := box.Stop(context.Background()); err != nil {
			t.Error(err)
		}
		fmt.Println("stop", time.Since(start))
	}
}

func TestErrorStates(t *testing.T) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}

	// Use the same pool to help catch tests related to atypical execution
	// order. Will possibly make tests harder to debug. My kingdom for
	// determinism.
	pool, err := boxpool.New(context.Background(), "runner", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Shutdown)

	for _, tt := range []struct {
		name   string
		doWork func(*testing.T, *boxpool.Pool)
	}{
		{
			name: "invalid datadir",
			doWork: func(t *testing.T, p *boxpool.Pool) {
				_, err := p.RunBox(context.Background(), nil, "/a/s/d/f/no-exist", nil)
				assert.Error(t, err)
			},
		},
		{
			name: "container killed",
			doWork: func(t *testing.T, p *boxpool.Pool) {
				box, err := p.RunBox(context.Background(), []string{"sleep", "10000000"}, t.TempDir(), nil)
				assert.NoError(t, err)
				if err := dockerClient.ContainerKill(context.Background(), box.ContainerID(), "SIGKILL"); err != nil {
					t.Fatal(err)
				}
				if _, _, err := box.Status(); err == nil {
					t.Fatal("expected error from box.Status, container has been killed")
				}
			},
		},
		{
			name: "stop twice",
			doWork: func(t *testing.T, p *boxpool.Pool) {
				start := time.Now()
				box, err := p.RunBox(context.Background(), []string{"sleep", "10000000"}, t.TempDir(), nil)
				assert.NoError(t, err)
				fmt.Println("runBox", time.Since(start))
				assert.NoError(t, err)
				if err := box.Stop(context.Background()); err != nil {
					t.Fatal(err)
				}
				if err := box.Stop(context.Background()); err == nil {
					t.Fatal("expected error from box.Stop, container already stopped")
				}
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.doWork(t, pool)

			states, err := boxpool.GetContainerStates(context.Background(), pool)
			if err != nil {
				t.Fatal(err)
			}
			spew.Dump(states)
			for _, state := range states {
				if state.InUse {
					t.Errorf("Container is in-use, but it shouldn't be: %v", state)
				}
				if state.Healthy && !state.State.Running {
					t.Errorf("Container is marked as healthy, but it is not running: %v", state)
				}
			}
		})
	}
}
