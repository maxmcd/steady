package execx_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/maxmcd/steady/internal/execx"
	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
	t.Run("shutdown with int", func(t *testing.T) {
		cmd := execx.Command("sleep", "100000")
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := cmd.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
	})
	t.Run("shutdown with kill", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := execx.Command("bash", "-c", "echo 'started' && trap \"\" INT && sleep 1000000")
		cmd.Stderr = &buf
		cmd.Stdout = &buf
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		for {
			// If bash doesn't have time to set the trap we'll kill instantly
			// with sigint
			if strings.Contains(buf.String(), "started") {
				break
			}
			time.Sleep(time.Millisecond * 1)
		}
		t.Log("cmd running, attempting to kill")
		start := time.Now()
		timeout := time.Millisecond * 10
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		if err := cmd.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
		require.GreaterOrEqual(t, time.Since(start), timeout)
		cancel()
		fmt.Println(buf.String())
	})
}
