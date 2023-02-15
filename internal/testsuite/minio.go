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
	"github.com/maxmcd/steady/internal/steadyutil"
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

func NewMinioServer(ctx context.Context, dir string) (*MinioServer, error) {
	start := time.Now()

	port, err := netx.GetFreePort()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	cmd := exec.CommandContext(ctx, "minio", "server", "--address="+addr, dir)
	cmd.Stderr = io.Discard
	cmd.Stdout = io.Discard

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	eg.Go(cmd.Wait)

	server := MinioServer{
		Address:    addr,
		Username:   "minioadmin",
		Password:   "minioadmin",
		BucketName: steadyutil.RandomString(15),
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

func (server *MinioServer) CycleBucket() error {
	s3Client, err := server.S3Client()
	if err != nil {
		return err
	}
	var innerErr error

	if err := s3Client.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(server.BucketName),
	}, func(loo *s3.ListObjectsOutput, b bool) bool {
		objects := []*s3.ObjectIdentifier{}
		if len(loo.Contents) == 0 {
			return false
		}
		for _, obj := range loo.Contents {
			objects = append(objects, &s3.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		if _, innerErr = s3Client.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: aws.String(server.BucketName),
			Delete: &s3.Delete{
				Objects: objects,
			},
		}); err != nil {
			return false
		}
		return true
	}); err != nil {
		return fmt.Errorf("error listing objects: %w", err)
	}
	if innerErr != nil {
		return fmt.Errorf("error deleting objects: %w", innerErr)
	}

	if _, err := s3Client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(server.BucketName),
	}); err != nil {
		return fmt.Errorf("error deleting bucket: %w", err)
	}

	server.BucketName = steadyutil.RandomString(15)

	if _, err := s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(server.BucketName),
	}); err != nil {
		return fmt.Errorf("error creating bucket: %w", err)
	}
	return nil
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
	return server.Wait()
}

func (server *MinioServer) Wait() error {
	if err := server.eg.Wait(); err != nil && !strings.Contains(err.Error(), "killed") {
		return err
	}
	return nil
}
