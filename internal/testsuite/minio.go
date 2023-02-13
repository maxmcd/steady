package testsuite

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
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

func NewMinioServer(dir string) (*MinioServer, error) {
	start := time.Now()

	port, err := netx.GetFreePort()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	cmd := exec.CommandContext(ctx, "minio", "server", "--address="+addr, dir)
	cmd.Stderr = io.Discard
	cmd.Stdout = io.Discard

	if err := cmd.Start(); err != nil {
		return nil, err
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

	s3Client, err := server.S3Client()
	if err != nil {
		return nil, err
	}

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
		return nil, err
	}
	fmt.Println("minio startup", time.Since(start))
	return &server, nil
}

func (server *MinioServer) S3Config() *aws.Config {
	return &aws.Config{
		Credentials:      credentials.NewStaticCredentials(server.Username, server.Password, ""),
		Endpoint:         aws.String("http://" + server.Address),
		Region:           aws.String("us-west-2"),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
	}
}

func (server *MinioServer) S3Client() (*s3.S3, error) {
	newSession, err := session.NewSession(server.S3Config())
	if err != nil {
		return nil, err
	}

	return s3.New(newSession), nil
}

func (server *MinioServer) Stop() error {
	server.cancel()
	if err := server.eg.Wait(); err != nil && !strings.Contains(err.Error(), "killed") {
		return err
	}
	return nil
}
