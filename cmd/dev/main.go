package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/maxmcd/steady/daemon"
	"github.com/maxmcd/steady/internal/testsuite"
	"github.com/maxmcd/steady/loadbalancer"
	"github.com/maxmcd/steady/slicer"
	"github.com/maxmcd/steady/steady"
	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/maxmcd/steady/web"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}

	for _, dir := range []string{"minio", "daemon-1", "daemon-2"} {
		_ = os.Mkdir(filepath.Join(tmpDir, dir), 0700)
	}

	minioServer, err := testsuite.NewMinioServer(filepath.Join(tmpDir, "minio"))
	if err != nil {
		panic(err)
	}

	daemonS3Config := daemon.S3Config{
		AccessKeyID:     minioServer.Username,
		SecretAccessKey: minioServer.Password,
		Bucket:          minioServer.BucketName,
		Endpoint:        "http://" + minioServer.Address,
		SkipVerify:      true,
		ForcePathStyle:  true,
	}

	assigner := slicer.Assigner{}

	daemon1 := daemon.NewDaemon(filepath.Join(tmpDir, "daemon-1"), ":8091", daemon.DaemonOptionWithS3(daemonS3Config))
	daemon1.Start(ctx)
	if err := assigner.AddHost(daemon1.ServerAddr(), nil); err != nil {
		panic(err)
	}

	daemon2 := daemon.NewDaemon(filepath.Join(tmpDir, "daemon-2"), ":8092", daemon.DaemonOptionWithS3(daemonS3Config))
	daemon2.Start(ctx)
	if err := assigner.AddHost(daemon2.ServerAddr(), nil); err != nil {
		panic(err)
	}

	lb := loadbalancer.NewLB()
	if err := lb.NewHostAssignments(assigner.Assignments()); err != nil {
		panic(err)
	}

	if err := lb.Start(ctx, ":8081", ":8082"); err != nil {
		panic(err)
	}

	steadyHandler :=
		steady.NewServer(
			steady.ServerOptions{
				PublicLoadBalancerURL:  "http://localhost:8081",
				PrivateLoadBalancerURL: "http://localhost:8082",
				DaemonClient:           daemon.NewClient("http://localhost:8082", nil),
			},
			steady.OptionWithSqlite("./steady.sqlite"))

	webHandler, err := web.NewServer(steadyrpc.NewSteadyProtobufClient("http://localhost:8080", http.DefaultClient))
	if err != nil {
		panic(err)
	}

	panic(
		http.ListenAndServe(":8080", web.WebAndSteadyHandler(steadyHandler, webHandler)),
	)
}
