package daemon_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/maxmcd/steady/daemon"
)

func (suite *DaemonSuite) TestLitestream() {
	t := suite.T()

	suite.StartMinioServer()

	d, _ := suite.CreateDaemon(daemon.DaemonOptionWithS3(suite.MinioServerS3Config()))

	client := suite.NewClient(d)
	app, err := client.CreateApplication(context.Background(), "max.db", suite.LoadExampleScript("http"))
	if err != nil {
		t.Fatal(err)
	}

	counter := 0
	createRecordRequest := func() {
		resp, err := http.Post(
			suite.DaemonURL(d, app.Name),
			"application/json", bytes.NewBuffer([]byte(`{"email":"lite"}`)))
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
	createRecordRequest()
	createRecordRequest()
	createRecordRequest()
	d.StopAllApplications()
	createRecordRequest()
	createRecordRequest()

	if _, err := client.DeleteApplication(context.Background(), app.Name); err != nil {
		t.Fatal(err)
	}

	app, err = client.CreateApplication(context.Background(), "max.db", suite.LoadExampleScript("http"))
	if err != nil {
		t.Fatal(err)
	}
	createRecordRequest()
}
