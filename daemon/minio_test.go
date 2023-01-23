package daemon

import (
	"fmt"
	"io"
	"os/exec"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type MinioServer struct {
	Username   string
	Password   string
	Address    string
	BucketName string
}

func NewMinioServer(t *testing.T) MinioServer {
	dir := t.TempDir()

	port, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}
	addr := fmt.Sprintf("localhost:%d", port)
	cmd := exec.Command("minio", "server", "--address="+addr, dir)
	cmd.Stderr = io.Discard
	cmd.Stdout = io.Discard

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			t.Fatal(err)
		}
	}()

	server := MinioServer{
		Address:    addr,
		Username:   "minioadmin",
		Password:   "minioadmin",
		BucketName: "litestream",
	}
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(server.Username, server.Password, ""),
		Endpoint:         aws.String("http://" + addr),
		Region:           aws.String("us-west-2"),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		t.Fatal(err)
	}

	s3Client := s3.New(newSession)

	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond * time.Duration(i*i))
		if _, err = s3Client.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(server.BucketName),
		}); err == nil {
			break
		}
	}
	if err != nil {
		t.Fatal(err)
	}

	return server
}

func TestMinioServer(t *testing.T) { _ = NewMinioServer(t) }
