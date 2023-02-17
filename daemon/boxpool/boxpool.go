// package boxpool implements a pool of docker containers that can be used to
// run applications. It handles maintaining the pool of images and managing logs
// and providing files to each running application.
package boxpool

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Pool struct {
	numPending int
	image      string

	dockerClient *client.Client
}

func NewPool(ctx context.Context, image string) (*Pool, error) {
	p := &Pool{image: image}
	var err error
	p.dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	if err := p.startContainer(ctx); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Pool) startContainer(ctx context.Context) error {
	cont, err := p.dockerClient.ContainerCreate(ctx, &container.Config{
		Image:       p.image,
		Cmd:         nil,
		AttachStdin: true,
		OpenStdin:   true,
		Tty:         false,
	}, &container.HostConfig{}, nil, nil, "")
	if err != nil {
		return err
	}

	if err := p.dockerClient.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	resp, err := p.dockerClient.ContainerAttach(ctx, cont.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Logs:   true,
	})
	if err != nil {
		return err
	}
	go stdcopy.StdCopy(os.Stdout, os.Stderr, resp.Reader)
	fmt.Fprintln(resp.Conn, "Hello I am doing")
	fmt.Fprintln(resp.Conn, "Hello I am doing")

	return nil
}

func (p *Pool) RunBox(cmd []string, dir string) (*Box, error) {
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
