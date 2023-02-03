package steady_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/maxmcd/steady/internal/daemontest"
	"github.com/maxmcd/steady/steady"
	"github.com/maxmcd/steady/steady/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestServer(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	server := &steady.Server{}
	steady.OptionWithSqlite(filepath.Join(tmpDir, "db.sqlite"))(server)

	var service *rpc.Service
	{
		resp, err := server.CreateService(ctx, &rpc.CreateServiceRequest{
			Name: "foo",
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "foo", resp.Service.Name)
		service = resp.Service
	}

	{
		resp, err := server.CreateServiceVersion(ctx, &rpc.CreateServiceVersionRequest{
			ServiceId: service.Id,
			Version:   "v1",
			Source:    "console.log('hi');",
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, service.Id, resp.ServiceVersion.ServiceId)
	}
}

type ServerSuite struct {
	daemontest.DaemonSuite
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerSuite))
}

func (suite *ServerSuite) TestCreateApplication() {

}
