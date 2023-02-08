package steady_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/maxmcd/steady/internal/testsuite"
	"github.com/maxmcd/steady/steady"
	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	server := &steady.Server{}
	steady.OptionWithSqlite(filepath.Join(tmpDir, "db.sqlite"))(server)

	var service *steadyrpc.Service
	{
		resp, err := server.CreateService(ctx, &steadyrpc.CreateServiceRequest{
			Name: "foo",
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "foo", resp.Service.Name)
		service = resp.Service
	}

	{
		resp, err := server.CreateServiceVersion(ctx, &steadyrpc.CreateServiceVersionRequest{
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
	testsuite.Suite
}

func TestServerSuite(t *testing.T) {
	testsuite.Run(t, new(ServerSuite))
}

func (suite *ServerSuite) TestCreateApplication() {

}
