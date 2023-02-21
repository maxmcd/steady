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

var ErrContainerStopped = fmt.Errorf("container has stopped")

type Pool struct {
	image   string
	dataDir string

	dockerClient *client.Client

	lock *sync.Mutex
	pool []*poolContainer
}

func New(ctx context.Context, image string, dataDir string) (*Pool, error) {
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
		// Runtime:        "runsc",
	}, nil, nil, "")
	if err != nil {
		return nil, err
	}

	cont := &poolContainer{id: c.ID, pool: p, dataDir: dataDir, lock: &sync.Mutex{}}

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
		_ = cont.shutdown(context.Background())
		return nil, err
	}
	info, err := p.dockerClient.ContainerInspect(ctx, cont.id)
	if err != nil {
		_ = cont.shutdown(context.Background())
		return nil, err
	}

	cont.ipAddress = info.NetworkSettings.IPAddress
	cont.attach = resp

	pr, pw := io.Pipe()
	go func() {
		_, _ = stdcopy.StdCopy(pw, os.Stderr, cont.attach.Reader)
		_ = pw.Close()
	}()

	cont.scanner = bufio.NewScanner(pr)

	return cont, nil
}
func (p *Pool) addContainer(ctx context.Context) (*poolContainer, error) {
	cont, err := p.startContainer(ctx)
	if err != nil {
		return nil, err
	}
	p.lock.Lock()
	p.pool = append(p.pool, cont)
	p.lock.Unlock()
	return cont, nil
}
func (p *Pool) nextContainer(ctx context.Context) (*poolContainer, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	free := []int{}
	for i, cont := range p.pool {
		if !cont.isRunning() {
			free = append(free, i)
		}
	}
	if len(free) == 0 {
		return p.addContainer(ctx)
	}
	if len(free) == 1 {
		go func() { _, _ = p.addContainer(context.Background()) }()
	}
	return p.pool[free[0]], nil
}

type poolContainer struct {
	dataDir               string
	appDataReturnLocation string
	pool                  *Pool
	id                    string
	ipAddress             string

	lock    *sync.Mutex
	attach  types.HijackedResponse
	scanner *bufio.Scanner

	running bool
}

type ContainerAction struct {
	Action string
	Cmd    []string `json:",omitempty"`
	Env    []string `json:",omitempty"`
}
type ContainerResponse struct {
	ExitCode *int  `json:",omitempty"`
	Running  *bool `json:",omitempty"`
	Err      string
}

func (pc *poolContainer) sendMsg(req ContainerAction) (resp ContainerResponse, err error) {
	if err := json.NewEncoder(pc.attach.Conn).Encode(req); err != nil {
		return ContainerResponse{}, err
	}
	if ok := pc.scanner.Scan(); !ok {
		return ContainerResponse{}, fmt.Errorf("container has stopped")
	}
	respString := pc.scanner.Text()
	if err := json.Unmarshal([]byte(respString), &resp); err != nil {
		return ContainerResponse{}, err
	}
	if resp.Err != "" {
		return ContainerResponse{}, fmt.Errorf(resp.Err)
	}
	return resp, nil
}

// TODO
// func (pc *poolContainer) exec(ctx context.Context) {
// 	pc.pool.dockerClient.ContainerExecCreate(ctx, pc.id, types.ExecConfig{})
// }

func (pc *poolContainer) isRunning() bool {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	return pc.running
}
func (pc *poolContainer) setRunning(v bool) {
	pc.lock.Lock()
	pc.running = v
	pc.lock.Unlock()
}

func (pc *poolContainer) run(ctx context.Context, cmd []string, dataDir string, env []string) error {
	pc.lock.Lock()
	if pc.running {
		pc.lock.Unlock()
		return fmt.Errorf("poolContainer is already running")
	}
	pc.running = true
	pc.lock.Unlock()

	// Mv datadir to ./app so that mount sees it as /opt/app
	if err := os.Rename(dataDir, filepath.Join(pc.dataDir, "app")); err != nil {
		pc.setRunning(false)
		return err
	}
	pc.appDataReturnLocation = dataDir

	_, err := pc.sendMsg(ContainerAction{
		Action: "run",
		Cmd:    cmd,
		Env:    env,
	})
	if err != nil {
		pc.setRunning(false)
		return err
	}
	return nil
}

func (pc *poolContainer) stop(ctx context.Context) error {
	pc.lock.Lock()
	if !pc.running {
		pc.lock.Unlock()
		return fmt.Errorf("poolContainer is not running and can't be stopped")
	}
	pc.lock.Unlock()
	_, sendErr := pc.sendMsg(ContainerAction{
		Action: "stop",
	})
	_ = os.Rename(filepath.Join(pc.dataDir, "app"), pc.appDataReturnLocation)
	pc.setRunning(false)
	if sendErr != nil {
		return sendErr
	}
	return nil
}

func (pc *poolContainer) status(ctx context.Context) (exitCode int, running bool, err error) {
	resp, err := pc.sendMsg(ContainerAction{Action: "status"})
	if err != nil {
		return 0, false, err
	}
	return *resp.ExitCode, *resp.Running, nil
}

func (pc *poolContainer) shutdown(ctx context.Context) error {
	return pc.pool.dockerClient.ContainerKill(ctx, pc.id, "SIGKILL")
}

func (p *Pool) RunBox(ctx context.Context, cmd []string, dataDir string, env []string) (*Box, error) {
	cont, err := p.nextContainer(ctx)
	if err != nil {
		return nil, err
	}
	if err := cont.run(ctx, cmd, dataDir, env); err != nil {
		return nil, err
	}
	return &Box{pool: p, cont: cont, dataDir: dataDir}, nil
}

type Box struct {
	cont    *poolContainer
	pool    *Pool
	dataDir string
}

func (b *Box) IPAddress() string {
	return b.cont.ipAddress
}

// Exec opens a shell session within the box.
func (b *Box) Exec() {

}

func (b *Box) Status(ctx context.Context) (exitCode int, running bool, err error) {
	return b.cont.status(ctx)
}

// Stop stops the program and frees the container back to the pool.
func (b *Box) Stop(ctx context.Context) error {
	return b.cont.stop(ctx)
}
