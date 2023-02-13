package testsuite

import (
	"testing"
)

func TestMinioServer(t *testing.T) {
	if _, err := NewMinioServer(t.TempDir()); err != nil {
		t.Fatal(err)
	}
}
