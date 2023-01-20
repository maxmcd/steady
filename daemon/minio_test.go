package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

type MinioServer struct {
	Username string
	Password string
	Address  string
}

func NewMinioServer(t *testing.T) MinioServer {
	dir := t.TempDir()

	port, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}
	addr := fmt.Sprintf("localhost:%d", port)
	cmd := exec.Command("minio", "server", "--address="+addr, dir)
	cmd.Env = []string{
		"MINIO_ROOT_USER=root",
		"MINIO_ROOT_PASSWORD=password",
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			panic(err)
		}
	}()

	return MinioServer{
		Address:  addr,
		Username: "root",
		Password: "password",
	}
}
