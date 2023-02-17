// package boxpool implements a pool of docker containers that can be used to
// run applications. It handles maintaining the pool of images and managing logs
// and providing files to each running application.
package boxpool

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Pool struct {
	image   string
	dataDir string

	dockerClient *client.Client

	lock *sync.Mutex
	pool []*poolContainer
}

func NewPool(ctx context.Context, image string, dataDir string) (*Pool, error) {
	p := &Pool{
		image:   image,
		dataDir: dataDir,
		lock:    &sync.Mutex{},
	}
	var err error
	p.dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	cont, err := p.startContainer(ctx)
	if err != nil {
		return nil, err
	}
	p.pool = []*poolContainer{cont}

	return p, nil
}

func (p *Pool) startContainer(ctx context.Context) (_ *poolContainer, err error) {
	dataDir, err := os.MkdirTemp(p.dataDir, "")
	if err != nil {
		return nil, fmt.Errorf("error making container data dir: %w", err)
	}

	c, err := p.dockerClient.ContainerCreate(ctx, &container.Config{
		Image:       p.image,
		Cmd:         nil,
		AttachStdin: true,
		OpenStdin:   true,
		Tty:         false,
	}, &container.HostConfig{
		ReadonlyRootfs: true,
		Binds:          []string{dataDir + ":/opt"},
	}, nil, nil, "")
	if err != nil {
		return nil, err
	}

	cont := &poolContainer{id: c.ID, pool: p, dataDir: dataDir}

	if err := p.dockerClient.ContainerStart(ctx, cont.id, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}
	resp, err := p.dockerClient.ContainerAttach(ctx, cont.id, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Logs:   true,
	})
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	go func() {
		_, _ = stdcopy.StdCopy(pw, os.Stderr, cont.attach.Reader)
		pw.Close()
	}()

	cont.attach = resp
	cont.scanner = bufio.NewScanner(pr)

	return cont, nil
}

type poolContainer struct {
	dataDir string
	pool    *Pool
	id      string

	lock    *sync.Mutex
	attach  types.HijackedResponse
	scanner *bufio.Scanner

	running bool
}

type ContainerAction struct {
	Action string
	Cmd    []string
	Env    []string
}
type ContainerResponse struct {
	Action string
}

func (pc *poolContainer) Run(ctx context.Context, cmd []string, dataDir string, env []string) error {
	// Mv datadir to ./app so that mount sees it as /opt/app
	if err := os.Rename(dataDir, filepath.Join(pc.dataDir, "app")); err != nil {
		return err
	}

	if err := json.NewEncoder(pc.attach.Conn).Encode(ContainerAction{
		Action: "run",
		Cmd:    cmd,
		Env:    env,
	}); err != nil {
		return err
	}
	if ok := pc.scanner.Scan(); !ok {
		return pc.scanner.Err()
	}
	respString := pc.scanner.Text()
	fmt.Println("got it", respString)
	var resp ContainerResponse
	if err := json.Unmarshal([]byte(pc.scanner.Text()), &resp); err != nil {
		return err
	}
	fmt.Println(resp)
	return nil
}

func (pc *poolContainer) Stop(ctx context.Context) error {
	return pc.pool.dockerClient.ContainerKill(ctx, pc.id, "SIGKILL")
}

func (p *Pool) RunBox(ctx context.Context, cmd []string, dataDir string, env []string) (*Box, error) {
	return nil, p.pool[0].Run(ctx, cmd, dataDir, env)
	return nil, nil
}

type Box struct {
	pool    *Pool
	dataDir string
}

// Exec opens a shell session within the box.
func (b *Box) Exec() {

}

// Close stops the program and frees the container back to the pool.
func (b *Box) Close() error {
	return nil
}

// Pool of images. New command runs the command within the image and mounts the
// OS directory into the expected location.
