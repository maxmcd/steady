package daemon

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
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

	createRecordRequest := func() {
		resp, err := http.Post(
			suite.daemonURL(app.Name),
			"application/json", bytes.NewBuffer([]byte(`{"email":"lite"}`)))
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("%q\n", string(b))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	fmt.Println(suite.d.dataDirectory)

	createRecordRequest()
	time.Sleep(time.Second)
	createRecordRequest()
	createRecordRequest()
	createRecordRequest()
	time.Sleep(time.Second)
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
