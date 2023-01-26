package steady

import (
	"fmt"
	"testing"

	"github.com/maxmcd/steady/internal/daemontest"
	"github.com/maxmcd/steady/slicer"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	daemontest.DaemonSuite
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
func (suite *TestSuite) TestDeploy() {
	t := suite.T()

	suite.StartMinioServer()
	assigner := &slicer.Assigner{}

	d, _ := suite.CreateDaemon()
	if err := assigner.AddLocation(d.ServerAddr(), nil); err != nil {
		t.Fatal(err)
	}

	d2, _ := suite.CreateDaemon()
	if err := assigner.AddLocation(d2.ServerAddr(), nil); err != nil {
		t.Fatal(err)
	}

	fmt.Println(assigner.Host("foo.max.max"))
	fmt.Println(assigner.Host("bar.max.max"))
}
