package main

import (
	"context"
	"testing"
)

func Test_run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wait := run(ctx, Config{
		dataDirectory:             t.TempDir(),
		databaseFile:              t.TempDir() + "/steady.sqlite",
		daemonOneAddress:          ":0",
		daemonTwoAddress:          ":0",
		publicLoadBalancerAddress: ":0",
		privateLoadBalanceAddress: ":0",
		webServerAddress:          "localhost:0",
	})
	cancel()
	if err := wait(); err != nil {
		t.Fatal(err)
	}
}
