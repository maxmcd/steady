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
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	_ "github.com/maxmcd/steady/internal/slogx"
	"github.com/maxmcd/steady/internal/steadyutil"
	"github.com/sourcegraph/conc/iter"
	"golang.org/x/exp/slog"
)

var ErrContainerStopped = fmt.Errorf("container has stopped")

type Pool struct {
	image   string
	dataDir string

	dockerClient *client.Client

	newContainerWG *sync.WaitGroup
	lock           *sync.Mutex
	pool           []*poolContainer

	running bool

	gid, uid int
}

func New(ctx context.Context, image string, dataDir string) (*Pool, error) {
	p := &Pool{
		image:          image,
		dataDir:        dataDir,
		newContainerWG: &sync.WaitGroup{},
		lock:           &sync.Mutex{},
		gid:            os.Getegid(),
		uid:            os.Getuid(),
		running:        true,
	}
	var err error
	p.dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	if _, err := p.addContainer(ctx); err != nil {
		return nil, err
	}

	return p, nil
}

type ContainerState struct {
	ID      string
	State   *types.ContainerState
	InUse   bool
	Healthy bool
}

func GetContainerStates(ctx context.Context, pool *Pool) (states []ContainerState, err error) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	for _, cont := range pool.pool {
		if cont == nil {
			continue
		}
		info, err := pool.dockerClient.ContainerInspect(ctx, cont.id)
		if err != nil {
			return nil, fmt.Errorf("container inspect on container %q: %w", cont.id, err)
		}
		states = append(states, ContainerState{
			ID: cont.id, State: info.State, InUse: cont.isInUse(), Healthy: cont.isHealthy()})
	}
	return states, nil
}

func (p *Pool) Shutdown() {
	p.newContainerWG.Wait()
	p.lock.Lock()
	p.running = false
	defer p.lock.Unlock()
	iter.ForEach(p.pool, func(c **poolContainer) {
		if *c == nil {
			return
		}
		(*c).shutdown(context.Background())
	})
	p.pool = nil
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
		// TODO: gvisor doesn't support bun :(
		// [pid    14] <... io_uring_setup resumed>}) = -1 ENOSYS (Function not implemented)
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
		cont.shutdown(context.Background())
		return nil, err
	}
	info, err := p.dockerClient.ContainerInspect(ctx, cont.id)
	if err != nil {
		cont.shutdown(context.Background())
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
	p.lock.Lock()
	p.pool = append(p.pool, nil)
	p.lock.Unlock()
	p.newContainerWG.Add(1)
	cont, err := p.startContainer(ctx)
	if err != nil {
		return nil, err
	}
	p.lock.Lock()
	if !p.running {
		cont.shutdown(context.Background())
		return nil, nil
	}
	for i, c := range p.pool {
		if c == nil {
			p.pool[i] = cont
			p.lock.Unlock()
			p.newContainerWG.Done()
			return cont, nil
		}
	}
	panic("unreachable" + fmt.Sprint(p.pool))
}
func (p *Pool) nextContainer(ctx context.Context) (*poolContainer, error) {
	p.lock.Lock()
	free := []int{}
	pending := 0
	temp := p.pool[:0]
	for i, cont := range p.pool {
		if cont == nil {
			temp = append(temp, cont)
			pending++
			continue
		}
		if !cont.isHealthy() {
			go cont.shutdown(context.Background())
			continue
		}
		if !cont.isInUse() {
			free = append(free, i)
		}
		temp = append(temp, cont)
	}
	p.pool = temp // Remove unhealthy containers
	if len(free) == 0 {
		p.lock.Unlock()
		return p.addContainer(ctx)
	}
	if len(free) == 1 && pending < 2 {
		go func() { _, _ = p.addContainer(context.Background()) }()
	}
	defer p.lock.Unlock()
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

	inUse          bool
	containerState *types.ContainerState
}

type Exec struct {
	Cmd []string `json:",omitempty"`
	Env []string `json:",omitempty"`
	Gid int
	Uid int
}

type ContainerAction struct {
	Action string
	Exec   *Exec
}
type ContainerResponse struct {
	ExitCode *int  `json:",omitempty"`
	Running  *bool `json:",omitempty"`
	Err      string
}

func (pc *poolContainer) sendMsg(req ContainerAction) (resp ContainerResponse, err error) {
	// Assume lock is acquired already
	if err := json.NewEncoder(pc.attach.Conn).Encode(req); err != nil {
		return ContainerResponse{}, err
	}
	if ok := pc.scanner.Scan(); !ok {
		pc.handleUnexpectedClose()
		return ContainerResponse{}, ErrContainerStopped
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

func (pc *poolContainer) handleUnexpectedClose() {
	// Assume lock is acquired already
	info, err := pc.pool.dockerClient.ContainerInspect(context.Background(), pc.id)
	if err != nil {
		fmt.Println("WARN: error fetching closed container state: ", err)
	}
	if err == nil {
		pc.containerState = info.State
	}
	pc.inUse = false
}

// TODO
// func (pc *poolContainer) exec(ctx context.Context) {
// 	pc.pool.dockerClient.ContainerExecCreate(ctx, pc.id, types.ExecConfig{})
// }

func (pc *poolContainer) isHealthy() bool {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	return pc.containerState == nil
}

func (pc *poolContainer) isInUse() bool {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	return pc.inUse
}

func (pc *poolContainer) run(ctx context.Context, cmd []string, dataDir string, env []string) error {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	if pc.inUse {
		return fmt.Errorf("poolContainer is already running")
	}
	pc.inUse = true

	// Mv datadir to ./app so that mount sees it as /opt/app
	if err := os.Rename(dataDir, filepath.Join(pc.dataDir, "app")); err != nil {
		pc.inUse = false
		return err
	}
	pc.appDataReturnLocation = dataDir

	_, err := pc.sendMsg(ContainerAction{
		Action: "run",
		Exec: &Exec{
			Cmd: cmd,
			Env: env,
			Gid: pc.pool.gid,
			Uid: pc.pool.uid,
		},
	})
	if err != nil {
		pc.inUse = false
		return err
	}
	return nil
}

type StopInfo struct {
	DataDir string
	LogFile string
}

func (pc *poolContainer) stop() (_ *StopInfo, err error) {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	if !pc.inUse {
		return nil, fmt.Errorf("poolContainer is not running and can't be stopped")
	}
	_, sendErr := pc.sendMsg(ContainerAction{
		Action: "stop",
	})
	logName := filepath.Join(pc.pool.dataDir, fmt.Sprintf("%d-%s.log", time.Now().UnixNano(), steadyutil.RandomString(4)))
	// TODO: if we can't do this we should report the error and also clean the
	// container up to be used again?
	_ = os.Rename(filepath.Join(pc.dataDir, "app"), pc.appDataReturnLocation)
	_ = os.Rename(filepath.Join(pc.dataDir, "log.log"), logName)
	// TODO: we need to ensure the container is healthy and can be used again
	// before setting inUse to false
	pc.inUse = false
	if sendErr != nil {
		return nil, sendErr
	}
	return &StopInfo{DataDir: pc.appDataReturnLocation, LogFile: logName}, nil
}

func (pc *poolContainer) status() (exitCode int, running bool, err error) {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	if !pc.inUse {
		return 0, false, ErrContainerStopped
	}
	resp, err := pc.sendMsg(ContainerAction{Action: "status"})
	if err != nil {
		return 0, false, err
	}
	return *resp.ExitCode, *resp.Running, nil
}

func (pc *poolContainer) shutdown(ctx context.Context) {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	pc.inUse = false
	_ = pc.pool.dockerClient.ContainerKill(ctx, pc.id, "SIGKILL")
	_ = pc.pool.dockerClient.ContainerRemove(ctx, pc.id, types.ContainerRemoveOptions{})

	slog.Debug("Killed", "id", pc.id)
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

func (b *Box) Status() (exitCode int, running bool, err error) {
	return b.cont.status()
}

func (b *Box) ContainerID() string {
	return b.cont.id
}

// Stop stops the program and frees the container back to the pool.
func (b *Box) Stop() (*StopInfo, error) {
	return b.cont.stop()
}
