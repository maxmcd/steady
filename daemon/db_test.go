package daemon_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/maxmcd/steady/daemon"
	"github.com/maxmcd/steady/internal/daemontest"
)

func (suite *DaemonSuite) TestLitestream() {
	t := suite.T()

	minioServer := daemontest.NewMinioServer(t)
	fmt.Println(minioServer.Address)

	d, _ := suite.CreateDaemon(daemon.DaemonOptionWithS3(daemon.S3Config{
		AccessKeyID:     minioServer.Username,
		SecretAccessKey: minioServer.Password,
		Bucket:          minioServer.BucketName,
		Endpoint:        "http://" + minioServer.Address,
		SkipVerify:      true,
		ForcePathStyle:  true,
	}))

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
