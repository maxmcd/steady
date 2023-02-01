package steadyservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/maxmcd/steady/internal/daemontest"
	"github.com/maxmcd/steady/loadbalancer"
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

	// Migrate job
	// Start job on daemon
	// send requests to it from the load balancer
	// add another host
	// migrate the job to another daemon
	// ensure all requests make it to a live job

	appName := "bar.max.max"
	httpClient := &http.Client{}

	lb := loadbalancer.NewLB(
		loadbalancer.OptionWithAppNameExtractor(
			loadbalancer.TestHeaderExtractor))

	suite.StartMinioServer()
	assigner := &slicer.Assigner{}

	d, _ := suite.CreateDaemon()
	if err := assigner.AddHost(d.ServerAddr(), nil); err != nil {
		t.Fatal(err)
	}

	if err := lb.NewHostAssignments(assigner.Assignments()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	lb.Start(ctx, ":0")

	{
		// Confirm the application currently returns a 404
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s", lb.ServerAddr()), nil)
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
	if _, err := dClient.CreateApplication(ctx, appName, suite.LoadExampleScript("http")); err != nil {
		t.Fatal(err)
	}

	counter := 0
	createRecordRequest := func() {
		req, err := http.NewRequest(http.MethodPost,
			fmt.Sprintf("http://%s", lb.ServerAddr()),
			bytes.NewBuffer([]byte(`{"email":"lite"}`)))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Host", appName)
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var jsonResponse map[string]interface{}
		suite.Require().NoError(json.NewDecoder(resp.Body).Decode(&jsonResponse))
		suite.Require().Equal(http.StatusOK, resp.StatusCode)

		// Here's the real test. Ensure that every time we make a request the ID
		// increments, even through deletes/restarts
		counter++
		suite.Require().Equal(counter, int(jsonResponse["id"].(float64)))
	}
	createRecordRequest()
	d.StopAllApplications()

	d2, _ := suite.CreateDaemon()
	if err := assigner.AddHost(d2.ServerAddr(), nil); err != nil {
		t.Fatal(err)
	}
	if err := lb.NewHostAssignments(assigner.Assignments()); err != nil {
		t.Fatal(err)
	}

	// How are we finding what to move?
	if _, err := dClient.DeleteApplication(ctx, appName); err != nil {
		t.Fatal(err)
	}

	d2Client := suite.NewClient(d)
	if _, err := d2Client.CreateApplication(ctx, appName, suite.LoadExampleScript("http")); err != nil {
		t.Fatal(err)
	}
	createRecordRequest()
}
