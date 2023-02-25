package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/maxmcd/steady/internal/boxpool"
	"github.com/maxmcd/steady/internal/netx"
)

func Test_bunRun(t *testing.T) {
	pool, err := boxpool.New(context.Background(), "runner", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Shutdown)
	tests := []struct {
		name    string
		script  string
		wantErr bool
	}{
		{"junk script", "asdfasdf", true},
		{"no server", "console.log('hi')", true},
		{"no server long running", "setTimeout(() => {}, 100_000)", true},
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
			port, err := netx.GetFreePort()
			if err != nil {
				t.Fatal(err)
			}
			box, err := bunRun(pool, dir, port, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("bunRun() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if box != nil {
				if _, err := box.Stop(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
