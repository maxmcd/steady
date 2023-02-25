package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/maxmcd/steady/daemon"
	_ "github.com/maxmcd/steady/internal/slogx"
	"github.com/maxmcd/steady/internal/testsuite"
	"github.com/maxmcd/steady/loadbalancer"
	"github.com/maxmcd/steady/slicer"
	"github.com/maxmcd/steady/steady"
	"github.com/maxmcd/steady/web"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

type Config struct {
	dataDirectory             string
	daemonOneAddress          string
	daemonTwoAddress          string
	publicLoadBalancerAddress string
	privateLoadBalanceAddress string
	webServerAddress          string
}

func main() {
	cfg := Config{}
	cfg.dataDirectory = ""
	cfg.daemonOneAddress = ":8091"
	cfg.daemonTwoAddress = ":8092"
	cfg.publicLoadBalancerAddress = ":8081"
	cfg.privateLoadBalanceAddress = ":8082"
	cfg.webServerAddress = "0.0.0.0:8080"

	flag.StringVar(&cfg.daemonOneAddress, "daemon-one-address", cfg.daemonOneAddress, "")
	flag.StringVar(&cfg.daemonTwoAddress, "daemon-two-address", cfg.daemonTwoAddress, "")
	flag.StringVar(&cfg.publicLoadBalancerAddress, "public-load-balancer-address", cfg.publicLoadBalancerAddress, "")
	flag.StringVar(&cfg.privateLoadBalanceAddress, "private-load-balancer-address", cfg.privateLoadBalanceAddress, "")
	flag.StringVar(&cfg.webServerAddress, "web-server-address", cfg.webServerAddress, "")

	// Will exit with -h flag
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigs
		slog.Info("got signal, shutting down", "signal", sig)
		cancel()
		sig = <-sigs
		slog.Info("got another signal, terminating", "signal", sig)
		os.Exit(1)
	}()
	// Run then wait
	if err := run(ctx, cfg)(); err != nil {
		panic(err)
	}
}

func run(ctx context.Context, cfg Config) func() error {
	eg, ctx := errgroup.WithContext(ctx)

	if cfg.dataDirectory == "" {
		tmpDir, err := os.MkdirTemp("", "")
		if err != nil {
			panic(err)
		}
		cfg.dataDirectory = tmpDir
	}

	for _, dir := range []string{"minio", "daemon-1", "daemon-2"} {
		if err := os.Mkdir(filepath.Join(cfg.dataDirectory, dir), 0700); err != nil {
			panic(err)
		}
	}

	minioServer, err := testsuite.NewMinioServer(ctx, filepath.Join(cfg.dataDirectory, "minio"))
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

	daemon1 := daemon.NewDaemon(filepath.Join(cfg.dataDirectory, "daemon-1"), cfg.daemonOneAddress,
		daemon.DaemonOptionWithS3(daemonS3Config))
	if err := daemon1.Start(ctx); err != nil {
		panic(err)
	}
	if err := assigner.AddHost(daemon1.ServerAddr(), nil); err != nil {
		panic(err)
	}

	daemon2 := daemon.NewDaemon(filepath.Join(cfg.dataDirectory, "daemon-2"), cfg.daemonTwoAddress,
		daemon.DaemonOptionWithS3(daemonS3Config))
	if err := daemon2.Start(ctx); err != nil {
		panic(err)
	}
	if err := assigner.AddHost(daemon2.ServerAddr(), nil); err != nil {
		panic(err)
	}

	lb := loadbalancer.NewLB()
	if err := lb.NewHostAssignments(assigner.Assignments()); err != nil {
		panic(err)
	}

	if err := lb.Start(ctx, cfg.publicLoadBalancerAddress, cfg.privateLoadBalanceAddress); err != nil {
		panic(err)
	}

	steadyHandler :=
		steady.NewServer(
			steady.ServerOptions{
				PublicLoadBalancerURL:  "http://" + lb.PublicServerAddr(),
				PrivateLoadBalancerURL: "http://" + lb.PrivateServerAddr(),
				DaemonClient:           daemon.NewClient("http://"+lb.PrivateServerAddr(), nil),
			},
			steady.OptionWithSqlite(filepath.Join(cfg.dataDirectory, "./steady.sqlite")))

	server, err := web.NewServer(steadyHandler)
	if err != nil {
		panic(err)
	}
	if err := server.Start(ctx, cfg.webServerAddress); err != nil {
		panic(err)
	}

	eg.Go(daemon1.Wait)
	eg.Go(daemon2.Wait)
	eg.Go(minioServer.Wait)
	eg.Go(lb.Wait)
	eg.Go(server.Wait)
	return eg.Wait
}
