package boxpool_test

import (
	"context"
	"testing"
	"time"

	"github.com/maxmcd/steady/daemon/boxpool"
)

func TestBasic(t *testing.T) {
	pool, err := boxpool.NewPool(context.Background(), "runner")
	if err != nil {
		panic(err)
	}
	_ = pool

	time.Sleep(time.Second)
}
