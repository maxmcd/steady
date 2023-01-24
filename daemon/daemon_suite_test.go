package daemon

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DaemonSuite struct {
	suite.Suite

	d      *Daemon
	cancel func()
	port   int
}

func TestDaemonSuite(t *testing.T) {
	// for i := 0; i < 100; i++ {
	suite.Run(t, new(DaemonSuite))
	// if t.Failed() {
	// 	return
	// }
	// }
}

var _ suite.BeforeTest = new(DaemonSuite)
var _ suite.AfterTest = new(DaemonSuite)

func (suite *DaemonSuite) BeforeTest(suiteName, testName string) {
	var err error
	suite.port, err = getFreePort()
	if err != nil {
		suite.T().Fatal(err)
	}
	suite.d = NewDaemon(suite.T().TempDir(), suite.port)
	var ctx context.Context
	ctx, suite.cancel = context.WithCancel(context.Background())
	suite.d.Start(ctx)
}

func (suite *DaemonSuite) AfterTest(suiteName, testName string) {
	suite.cancel()
	suite.cancel = nil
	if err := suite.d.Wait(); err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *DaemonSuite) daemonURL(paths ...string) string {
	return fmt.Sprintf("http://localhost:%d", suite.port) + filepath.Join(append([]string{"/"}, paths...)...)
}

func (suite *DaemonSuite) newClient() *Client {
	client, err := NewClient(suite.daemonURL())
	suite.Require().NoError(err)
	return client
}

func (suite *DaemonSuite) loadExampleScript(name string) string {
	abs, err := filepath.Abs("../examples/" + name)
	if err != nil {
		suite.T().Fatal(err)
	}
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		suite.T().Fatalf("example script %q does not exist", name)
	}
	f, err := os.Open(filepath.Join(abs, "index.ts"))
	suite.Require().NoError(err)
	b, err := io.ReadAll(f)
	suite.Require().NoError(err)
	return string(b)
}
