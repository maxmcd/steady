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

	"github.com/maxmcd/steady/daemon"
	"github.com/maxmcd/steady/loadbalancer"
	"github.com/maxmcd/steady/slicer"
	"github.com/maxmcd/steady/steady"
	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/maxmcd/steady/web"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite

	daemons       []*daemon.Daemon
	assigner      *slicer.Assigner
	loadBalancers []*loadbalancer.LB

	cancels []func()

	minioEnabled bool
	minioServer  *MinioServer
}

// NewDaemon creates a daemon with the provided options. If you've called
// StartMinioServer, that server will be associated with the created Daemon.
func (suite *Suite) NewDaemon(opts ...daemon.DaemonOption) (d *daemon.Daemon, dir string) {
	dir = suite.T().TempDir()

	fmt.Println("HUH!", suite.minioEnabled)
	if suite.minioEnabled {
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
	t := suite.T()
	for _, cancel := range suite.cancels {
		cancel()
	}
	suite.cancels = nil
	for _, daemon := range suite.daemons {
		if err := daemon.Wait(); err != nil {
			t.Error(err)
		}
	}
	suite.daemons = nil
	for _, lb := range suite.loadBalancers {
		if err := lb.Wait(); err != nil {
			t.Error(err)
		}
	}
	suite.loadBalancers = nil
	if suite.minioEnabled {
		if err := suite.minioServer.CycleBucket(); err != nil {
			t.Fatal(err)
		}
	}
	suite.minioEnabled = false
}

func (suite *Suite) NewSteadyServer() http.Handler {
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

type EmailSink struct {
	Emails []string
}

func (es *EmailSink) LatestEmail() string {
	if len(es.Emails) == 0 {
		return ""
	}
	return es.Emails[len(es.Emails)-1]
}

func (suite *Suite) NewWebServer(opts steady.ServerOptions) (*EmailSink, string) {
	es := &EmailSink{}
	sqliteDataDir := suite.T().TempDir()
	steadyHandler := steady.NewServer(
		opts,
		steady.OptionWithSqlite(filepath.Join(sqliteDataDir, "./steady.sqlite")),
		steady.OptionWithEmailSink(func(email string) {
			es.Emails = append(es.Emails, email)
		}),
	)
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
	server := http.Server{Handler: web.WebAndSteadyHandler(steadyHandler, webHandler)}
	ctx, cancel := context.WithCancel(context.Background())
	suite.cancels = append(suite.cancels, cancel)
	go func() { _ = server.Serve(listener) }()
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	return es, url
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
	if err := lb.Start(ctx, "localhost:0", "localhost:0"); err != nil {
		suite.T().Fatal(err)
	}
	suite.cancels = append(suite.cancels, cancel)
	suite.loadBalancers = append(suite.loadBalancers, lb)
	return lb
}

func (suite *Suite) StartMinioServer() {
	if suite.minioEnabled && suite.minioServer != nil {
		return
	}
	suite.minioEnabled = true
	if suite.minioServer != nil {
		return
	}
	var err error
	suite.minioServer, err = NewMinioServer(suite.T().TempDir())
	if err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *Suite) MinioServerS3Config() daemon.S3Config {
	if !suite.minioEnabled {
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

func (suite *Suite) NewDaemonClient(addr string) *daemon.Client {
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
