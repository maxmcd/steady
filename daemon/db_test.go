package daemon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (suite *DaemonSuite) TestLitestream() {
	t := suite.T()

	minIOServer := NewMinioServer(t)
	fmt.Println(minIOServer.Address)

	suite.d.s3Config = &S3Config{
		AccessKeyID:     minIOServer.Username,
		SecretAccessKey: minIOServer.Password,
		Bucket:          minIOServer.BucketName,
		Endpoint:        "http://" + minIOServer.Address,
		SkipVerify:      true,
		ForcePathStyle:  true,
	}

	client := suite.newClient()
	app, err := client.CreateApplication("max.db", suite.loadExampleScript("http"))
	if err != nil {
		t.Fatal(err)
	}

	counter := 0
	createRecordRequest := func() {
		resp, err := http.Post(
			suite.daemonURL(app.Name),
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
	fmt.Println(suite.d.dataDirectory)

	createRecordRequest()
	suite.d.StopAllApplications()
	createRecordRequest()
	createRecordRequest()
	createRecordRequest()
	suite.d.StopAllApplications()
	createRecordRequest()
	createRecordRequest()

	if _, err := client.DeleteApplication(app.Name); err != nil {
		t.Fatal(err)
	}
	fmt.Println("deleted")

	app, err = client.CreateApplication("max.db", suite.loadExampleScript("http"))
	if err != nil {
		t.Fatal(err)
	}
	createRecordRequest()
}
