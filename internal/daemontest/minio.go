package daemontest

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
	"github.com/maxmcd/steady/internal/netx"
	"golang.org/x/sync/errgroup"
)

type MinioServer struct {
	Username   string
	Password   string
	Address    string
	BucketName string

	cancel func()
	eg     *errgroup.Group
}

func NewMinioServer(t *testing.T) *MinioServer {
	start := time.Now()

	dir := t.TempDir()

	port, err := netx.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	cmd := exec.CommandContext(ctx, "minio", "server", "--address="+addr, dir)
	cmd.Stderr = io.Discard
	cmd.Stdout = io.Discard

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	eg.Go(cmd.Wait)

	server := MinioServer{
		Address:    addr,
		Username:   "minioadmin",
		Password:   "minioadmin",
		BucketName: "litestream",
		eg:         eg,
		cancel:     cancel,
	}

	s3Client := server.s3Client(t)

	for i := 0; i < 10; i++ {
		fmt.Println("minio startup", time.Since(start))
		time.Sleep(time.Millisecond * time.Duration(i*i))
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
		if _, err = s3Client.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(server.BucketName),
		}); err == nil {
			cancel()
			break
		} else if strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") {
			err = nil
			cancel()
			break
		}
		cancel()
	}
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("minio startup", time.Since(start))
	return &server
}

func (server *MinioServer) s3Client(t *testing.T) *s3.S3 {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(server.Username, server.Password, ""),
		Endpoint:         aws.String("http://" + server.Address),
		Region:           aws.String("us-west-2"),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		t.Fatal(err)
	}

	return s3.New(newSession)
}

func (server *MinioServer) Stop(t *testing.T) {
	server.cancel()
	if err := server.eg.Wait(); err != nil && !strings.Contains(err.Error(), "killed") {
		t.Error(err)
	}
}
