package daemontest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/maxmcd/steady/daemon"
	"github.com/stretchr/testify/suite"
)

type DaemonSuite struct {
	suite.Suite

	daemons []*daemon.Daemon
	cancels []func()

	minioServer *MinioServer
}

var _ suite.SetupAllSuite = new(DaemonSuite)
var _ suite.BeforeTest = new(DaemonSuite)
var _ suite.AfterTest = new(DaemonSuite)

func (suite *DaemonSuite) SetupSuite() {}

// CreateDaemon creates a daemon with the provided options. If you've called
// StartMinioServer, that server will be associated with the created Daemon.
func (suite *DaemonSuite) CreateDaemon(opts ...daemon.DaemonOption) (d *daemon.Daemon, dir string) {
	dir = suite.T().TempDir()

	if suite.minioServer != nil {
		opts = append(opts, daemon.DaemonOptionWithS3(daemon.S3Config{
			AccessKeyID:     suite.minioServer.Username,
			SecretAccessKey: suite.minioServer.Password,
			Bucket:          suite.minioServer.BucketName,
			Endpoint:        "http://" + suite.minioServer.Address,
			SkipVerify:      true,
			ForcePathStyle:  true,
		}))
	}
	d = daemon.NewDaemon(dir, "localhost:0", "localhost:0", opts...)
	ctx, cancel := context.WithCancel(context.Background())
	d.Start(ctx)
	suite.cancels = append(suite.cancels, cancel)
	suite.daemons = append(suite.daemons, d)
	return d, dir
}

func (suite *DaemonSuite) StartMinioServer() {
	suite.minioServer = NewMinioServer(suite.T())
}

func (suite *DaemonSuite) MinioServerS3Config() daemon.S3Config {
	if suite.minioServer == nil {
		suite.T().Fatal("must call StartMinioServer before this method")
	}
	return daemon.S3Config{
		AccessKeyID:     suite.minioServer.Username,
		SecretAccessKey: suite.minioServer.Password,
		Bucket:          suite.minioServer.BucketName,
		Endpoint:        "http://" + suite.minioServer.Address,
		SkipVerify:      true,
		ForcePathStyle:  true,
	}
}

func (suite *DaemonSuite) BeforeTest(suiteName, testName string) {}

func (suite *DaemonSuite) AfterTest(suiteName, testName string) {
	for _, cancel := range suite.cancels {
		cancel()
	}
	for _, daemon := range suite.daemons {
		if err := daemon.Wait(); err != nil {
			suite.T().Error(err)
		}
	}
	if suite.minioServer != nil {
		suite.minioServer.Stop(suite.T())
		suite.minioServer = nil
	}
}

func (suite *DaemonSuite) Request(
	d *daemon.Daemon, appName string,
	method string, url string,
	body string) (_ *http.Response, respBody string, err error) {
	req, err := http.NewRequest(method, suite.DaemonURL(d, url), bytes.NewBufferString(body))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Host", appName+".foo.com")
	req.Host = appName + ".foo.com"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, "", err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, "", err
	}
	return resp, string(b), err
}
func (suite *DaemonSuite) DaemonURL(d *daemon.Daemon, paths ...string) string {
	return fmt.Sprintf("http://"+d.PublicServerAddr()) + filepath.Join(append([]string{"/"}, paths...)...)
}

func (suite *DaemonSuite) NewClient(d *daemon.Daemon) daemon.Client {
	return daemon.NewClient(fmt.Sprintf("http://"+d.PrivateServerAddr()), nil)
}

func (suite *DaemonSuite) LoadExampleScript(name string) string {
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
