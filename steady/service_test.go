package steady_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/maxmcd/steady/steady"
	"github.com/maxmcd/steady/steady/rpc"
	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	tmpDir := t.TempDir()
	service := &steady.Service{}
	steady.OptionWithSqlite(filepath.Join(tmpDir, "db.sqlite"))(service)

	resp, err := service.CreateService(context.Background(), &rpc.CreateServiceRequest{
		Name: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "foo", resp.Service.Name)
}
