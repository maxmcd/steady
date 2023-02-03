package steady

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	daemonrpc "github.com/maxmcd/steady/daemon/rpc"
	"github.com/maxmcd/steady/internal/daemontest"
	"github.com/maxmcd/steady/steady/rpc"
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

	// Migrate job
	// Start job on daemon
	// send requests to it from the load balancer
	// add another host
	// migrate the job to another daemon
	// ensure all requests make it to a live job
	appName := "whee"
	httpClient := &http.Client{}

	suite.StartMinioServer()
	d, _ := suite.NewDaemon()
	lb := suite.NewLB()
	ctx := context.Background()
	{
		// Confirm the application currently returns a 404
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s", lb.PublicServerAddr()), nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Host", appName)
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		suite.Equal(http.StatusNotFound, resp.StatusCode)
	}

	dClient := suite.NewClient(d)
	if _, err := dClient.CreateApplication(ctx, &daemonrpc.CreateApplicationRequest{
		Name:   appName,
		Script: suite.LoadExampleScript("http"),
	}); err != nil {
		t.Fatal(err)
	}

	counter := 0
	createRecordRequest := func() {
		req, err := http.NewRequest(http.MethodPost,
			fmt.Sprintf("http://%s", lb.PublicServerAddr()),
			bytes.NewBuffer([]byte(`{"email":"lite"}`)))
		if err != nil {
			t.Fatal(err)
		}
		req.Host = appName
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		suite.Require().Equal(http.StatusOK, resp.StatusCode)
		var jsonResponse map[string]interface{}
		suite.Require().NoError(json.NewDecoder(resp.Body).Decode(&jsonResponse))

		// Here's the real test. Ensure that every time we make a request the ID
		// increments, even through deletes/restarts
		counter++
		suite.Require().Equal(counter, int(jsonResponse["id"].(float64)))
	}
	createRecordRequest()
	d.StopAllApplications()

	d2, _ := suite.NewDaemon()

	// How are we finding what to move?
	if _, err := dClient.DeleteApplication(ctx, &daemonrpc.DeleteApplicationRequest{Name: appName}); err != nil {
		t.Fatal(err)
	}

	d2Client := suite.NewClient(d2)
	if _, err := d2Client.CreateApplication(ctx, &daemonrpc.CreateApplicationRequest{
		Name:   appName,
		Script: suite.LoadExampleScript("http"),
	}); err != nil {
		t.Fatal(err)
	}
	createRecordRequest()
}

func (suite *TestSuite) TestServer() {
	t := suite.T()

	// Migrate job
	// Start job on daemon
	// send requests to it from the load balancer
	// add another host
	// migrate the job to another daemon
	// ensure all requests make it to a live job

	suite.StartMinioServer()

	ctx := context.Background()

	d, _ := suite.NewDaemon()

	lb := suite.NewLB()

	server := NewServer(ServerOptions{
		PrivateLoadBalancerURL: lb.PrivateServerAddr(),
		PublicLoadBalancerURL:  lb.PublicServerAddr(),
		DaemonClient:           suite.NewClient(d),
	}, OptionWithSqlite(t.TempDir()+"/foo.sqlite"))

	resp, err := server.DeploySource(ctx, &rpc.DeploySourceRequest{
		Source: suite.LoadExampleScript("http"),
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp.Url)
	{
		req, err := http.NewRequest(http.MethodGet, "http://"+lb.PublicServerAddr(), nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Host = resp.Url

		httpResp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		suite.Equal(http.StatusOK, httpResp.StatusCode)
		_, _ = io.Copy(os.Stdout, httpResp.Body)
	}
}
