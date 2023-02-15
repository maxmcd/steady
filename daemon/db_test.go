package daemon_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/maxmcd/steady/daemon"
	"github.com/maxmcd/steady/daemon/daemonrpc"
)

func (suite *DaemonSuite) TestLitestream() {
	t := suite.T()

	suite.StartMinioServer()

	d, _ := suite.NewDaemon(daemon.DaemonOptionWithS3(suite.MinioServerS3Config()))

	client := suite.NewDaemonClient(d.ServerAddr())
	app, err := client.CreateApplication(context.Background(), &daemonrpc.CreateApplicationRequest{
		Name:   "max-db",
		Script: suite.LoadExampleScript("http"),
	})
	if err != nil {
		t.Fatal(err)
	}

	counter := 0
	createRecordRequest := func() {
		resp, respBody, err := suite.DaemonRequest(d, app.Name, http.MethodPost, "/", `{"email":"lite"}`)
		if err != nil {
			t.Fatal(err)
		}
		suite.Require().Equal(http.StatusOK, resp.StatusCode)
		var jsonResponse map[string]interface{}
		fmt.Println(respBody)
		suite.Require().NoError(json.NewDecoder(bytes.NewBufferString(respBody)).Decode(&jsonResponse))

		// Here's the real test. Ensure that every time we make a request the ID
		// increments, even through deletes/restarts
		counter++
		suite.Require().Equal(counter, int(jsonResponse["id"].(float64)))
	}

	createRecordRequest()
	d.StopAllApplications()
	createRecordRequest()
	createRecordRequest()
	createRecordRequest()
	d.StopAllApplications()
	createRecordRequest()
	createRecordRequest()

	if _, err := client.DeleteApplication(context.Background(), &daemonrpc.DeleteApplicationRequest{
		Name: app.Name,
	}); err != nil {
		t.Fatal(err)
	}

	app, err = client.CreateApplication(context.Background(), &daemonrpc.CreateApplicationRequest{
		Name:   "max-db",
		Script: suite.LoadExampleScript("http"),
	})
	if err != nil {
		t.Fatal(err)
	}
	createRecordRequest()
}
