package testsuite

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/handlers"
	"github.com/maxmcd/steady/daemon"
	"github.com/maxmcd/steady/loadbalancer"
	"github.com/maxmcd/steady/slicer"
	"github.com/maxmcd/steady/steady"
	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/maxmcd/steady/web"
	"github.com/stretchr/testify/suite"
)

func Run(t *testing.T, s suite.TestingSuite) {
	countString := os.Getenv("STEADY_SUITE_RUN_COUNT")
	count := 1
	if countString != "" {
		var err error
		count, err = strconv.Atoi(countString)
		if err != nil {
			t.Fatal("failed to convert env var to int", err, countString)
		}
	}
	for i := 0; i < count; i++ {
		suite.Run(t, s)
		if t.Failed() {
			return
		}
	}
}

type Suite struct {
	suite.Suite

	daemons       []*daemon.Daemon
	assigner      *slicer.Assigner
	loadBalancers []*loadbalancer.LB

	cancels []func()

	minioServer *MinioServer
}

// NewDaemon creates a daemon with the provided options. If you've called
// StartMinioServer, that server will be associated with the created Daemon.
func (suite *Suite) NewDaemon(opts ...daemon.DaemonOption) (d *daemon.Daemon, dir string) {
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
	d = daemon.NewDaemon(dir, "localhost:0", opts...)
	ctx, cancel := context.WithCancel(context.Background())
	d.Start(ctx)
	suite.cancels = append(suite.cancels, cancel)
	suite.daemons = append(suite.daemons, d)
	if err := suite.assigner.AddHost(d.ServerAddr(), nil); err != nil {
		suite.T().Fatal(err)
	}
	newAssignments := suite.assigner.Assignments()
	for _, lb := range suite.loadBalancers {
		if err := lb.NewHostAssignments(newAssignments); err != nil {
			suite.T().Fatal(err)
		}
	}
	return d, dir
}

var _ suite.SetupAllSuite = new(Suite)
var _ suite.BeforeTest = new(Suite)
var _ suite.AfterTest = new(Suite)

func (suite *Suite) SetupSuite() {}

func (suite *Suite) BeforeTest(suiteName, testName string) {
	suite.assigner = &slicer.Assigner{}
}

func (suite *Suite) AfterTest(suiteName, testName string) {
	for _, cancel := range suite.cancels {
		cancel()
	}
	suite.cancels = nil
	for _, daemon := range suite.daemons {
		if err := daemon.Wait(); err != nil {
			suite.T().Error(err)
		}
	}
	suite.daemons = nil
	for _, lb := range suite.loadBalancers {
		if err := lb.Wait(); err != nil {
			suite.T().Error(err)
		}
	}
	suite.loadBalancers = nil
	if suite.minioServer != nil {
		suite.minioServer.Stop(suite.T())
		suite.minioServer = nil
	}
}

func (suite *Suite) NewSteadyServer() *steady.Server {
	t := suite.T()
	if len(suite.loadBalancers) == 0 {
		t.Fatal("need at least one load balancer to create a steady server")
	}
	return steady.NewServer(steady.ServerOptions{
		PrivateLoadBalancerURL: suite.loadBalancers[0].PrivateServerAddr(),
		PublicLoadBalancerURL:  suite.loadBalancers[0].PublicServerAddr(),
		DaemonClient:           suite.NewDaemonClient(suite.loadBalancers[0].PrivateServerAddr()),
	}, steady.OptionWithSqlite(t.TempDir()+"/steady.sqlite"))
}

func (suite *Suite) NewWebServer() string {
	sqliteDataDir := suite.T().TempDir()
	steadyHandler := steadyrpc.NewSteadyServer(
		steady.NewServer(
			steady.ServerOptions{},
			steady.OptionWithSqlite(filepath.Join(sqliteDataDir, "./steady.sqlite"))))
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		suite.T().Fatal(err)
	}
	url := fmt.Sprintf("http://%s", listener.Addr().String())
	webHandler, err := web.NewServer(
		steadyrpc.NewSteadyProtobufClient(
			url,
			http.DefaultClient))
	if err != nil {
		suite.T().Fatal(err)
	}
	server := http.Server{Handler: handlers.LoggingHandler(os.Stdout,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/twirp") {
				steadyHandler.ServeHTTP(w, r)
			} else {
				webHandler.ServeHTTP(w, r)
			}
		}),
	)}
	ctx, cancel := context.WithCancel(context.Background())
	suite.cancels = append(suite.cancels, cancel)
	go func() { _ = server.Serve(listener) }()
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	return url
}

func (suite *Suite) NewLB() *loadbalancer.LB {
	if len(suite.daemons) == 0 {
		suite.T().Fatal("You cannot create a load balancer if no daemon servers exis")
	}
	lb := loadbalancer.NewLB()
	if err := lb.NewHostAssignments(suite.assigner.Assignments()); err != nil {
		suite.T().Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := lb.Start(ctx, ":0", ":0"); err != nil {
		suite.T().Fatal(err)
	}
	suite.cancels = append(suite.cancels, cancel)
	suite.loadBalancers = append(suite.loadBalancers, lb)
	return lb
}

func (suite *Suite) StartMinioServer() {
	suite.minioServer = NewMinioServer(suite.T())
}

func (suite *Suite) MinioServerS3Config() daemon.S3Config {
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

func (suite *Suite) DaemonRequest(
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
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, "", err
	}
	return resp, string(b), err
}
func (suite *Suite) DaemonURL(d *daemon.Daemon, paths ...string) string {
	return fmt.Sprintf("http://"+d.ServerAddr()) + filepath.Join(append([]string{"/"}, paths...)...)
}

func (suite *Suite) NewDaemonClient(addr string) daemon.Client {
	return daemon.NewClient(fmt.Sprintf("http://%s", addr), nil)
}

func (suite *Suite) LoadExampleScript(name string) string {
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
