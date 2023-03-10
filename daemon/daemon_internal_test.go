package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/maxmcd/steady/internal/netx"
)

func Test_bunRun(t *testing.T) {
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
			cmd, err := bunRun(dir, port, nil, os.Stdout)
			fmt.Println(err)
			if (err != nil) != tt.wantErr {
				t.Errorf("bunRun() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cmd != nil {
				if err := cmd.Shutdown(context.TODO()); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
