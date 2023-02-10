package steady_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	daemonrpc "github.com/maxmcd/steady/daemon/daemonrpc"
	"github.com/maxmcd/steady/internal/testsuite"
)

type TestSuite struct {
	testsuite.Suite
}

func TestTestSuite(t *testing.T) {
	testsuite.Run(t, new(TestSuite))
}

func (suite *TestSuite) TestMigrate() {
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

	dClient := suite.NewDaemonClient(d.ServerAddr())
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
		var jsonResponse map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&jsonResponse)

		suite.Require().Equal(http.StatusOK, resp.StatusCode, jsonResponse)

		// Here's the real test. Ensure that every time we make a request the ID
		// increments, even through deletes/restarts
		counter++
		suite.Require().Equal(counter, int(jsonResponse["id"].(float64)), jsonResponse)
	}
	createRecordRequest()
	d.StopAllApplications()

	d2, _ := suite.NewDaemon()
	d2Client := suite.NewDaemonClient(d2.ServerAddr())

	// How are we finding what to move?
	if _, err := dClient.DeleteApplication(ctx, &daemonrpc.DeleteApplicationRequest{Name: appName}); err != nil {
		t.Fatal(err)
	}

	if _, err := d2Client.CreateApplication(ctx, &daemonrpc.CreateApplicationRequest{
		Name:   appName,
		Script: suite.LoadExampleScript("http"),
	}); err != nil {
		t.Fatal(err)
	}
	createRecordRequest()
}

// func (suite *TestSuite) TestServer() {
// 	t := suite.T()

// 	// Migrate job
// 	// Start job on daemon
// 	// send requests to it from the load balancer
// 	// add another host
// 	// migrate the job to another daemon
// 	// ensure all requests make it to a live job

// 	suite.StartMinioServer()

// 	ctx := context.Background()

// 	suite.NewDaemon()
// 	lb := suite.NewLB()
// 	server := suite.NewSteadyServer()
// 	resp, err := server.DeploySource(ctx, &steadyrpc.DeploySourceRequest{
// 		Source: suite.LoadExampleScript("http"),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	{
// 		req, err := http.NewRequest(http.MethodGet, "http://"+lb.PublicServerAddr(), nil)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		req.Host = resp.Url

// 		httpResp, err := http.DefaultClient.Do(req)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		suite.Equal(http.StatusOK, httpResp.StatusCode)
// 		_, _ = io.Copy(os.Stdout, httpResp.Body)
// 	}
// }
