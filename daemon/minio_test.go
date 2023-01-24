package daemon

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
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
	start := time.Now()

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

	go func() { _ = cmd.Wait() }()
	t.Cleanup(func() { _ = cmd.Process.Kill() })

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

	for i := 0; i < 10; i++ {
		fmt.Println("minio startup", time.Since(start))
		time.Sleep(time.Millisecond * time.Duration(i*i))
		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*50)
		if _, err = s3Client.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(server.BucketName),
		}); err == nil {
			break
		} else if err != nil && strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") {
			err = nil
			break
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("minio startup", time.Since(start))
	return server
}

func TestMinioServer(t *testing.T) { _ = NewMinioServer(t) }
