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

	"github.com/maxmcd/steady/daemon/boxpool"
	"github.com/maxmcd/steady/internal/steadyutil"
	"github.com/maxmcd/steady/internal/testsuite"
)

func TestBasic(t *testing.T) {
	dataDir := t.TempDir()

	suite := new(testsuite.Suite)
	suite.SetT(t)
	pool, err := boxpool.New(context.Background(), "runner", dataDir)
	if err != nil {
		t.Fatal(err)
	}

	appDir := t.TempDir()
	f, err := os.Create(filepath.Join(appDir, "index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(suite.LoadExampleScript("http"))
	f.Close()

	for i := 0; i < 5; i++ {
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

		for i := 0; i < 20; i++ {
			res, err := http.Get(fmt.Sprintf("http://%s:3000/"+healthEndpoint, box.IPAddress()))
			if err == nil {
				io.Copy(os.Stdout, res.Body)
				_ = res.Body.Close()
				break
			}
			exponent := time.Duration((i+1)*(i+1)) / 2
			time.Sleep(time.Millisecond * exponent)
			_, running, err := box.Status(context.Background())
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
