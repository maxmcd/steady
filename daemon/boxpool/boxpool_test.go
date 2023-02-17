package boxpool_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maxmcd/steady/daemon/boxpool"
	"github.com/maxmcd/steady/internal/testsuite"
)

func TestBasic(t *testing.T) {
	dataDir := t.TempDir()

	suite := new(testsuite.Suite)
	suite.SetT(t)
	pool, err := boxpool.NewPool(context.Background(), "runner", dataDir)
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

	fmt.Println(appDir, dataDir)
	if _, err := pool.RunBox(context.Background(),
		[]string{"bun", "run", "index.ts", "--no-install"},
		appDir, nil); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Hour)
}
