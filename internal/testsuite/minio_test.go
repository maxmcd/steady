package testsuite

import (
	"context"
	"testing"
)

func TestMinioServer(t *testing.T) {
	if _, err := NewMinioServer(context.Background(), t.TempDir()); err != nil {
		t.Fatal(err)
	}
}
