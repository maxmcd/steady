package execx

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type Cmd struct {
	exec.Cmd

	lock         *sync.RWMutex
	processState *os.ProcessState
	waitErr      error
	doneChan     chan struct{}
	started      bool
}

func Command(name string, arg ...string) *Cmd {
	return &Cmd{
		Cmd:      *exec.Command(name, arg...),
		doneChan: make(chan struct{}, 1),
		lock:     &sync.RWMutex{},
	}
}

func (c *Cmd) Start() error {
	if c.Cmd.SysProcAttr == nil {
		c.Cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// Attempt to kill all children when we fall back to sigkill
	// https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
	c.Cmd.SysProcAttr.Setpgid = true

	err := c.Cmd.Start()
	if err != nil {
		return err
	}
	c.lock.Lock()
	c.started = true
	c.lock.Unlock()
	go func() {
		err := c.Cmd.Wait()
		c.lock.Lock()
		c.processState = c.Cmd.ProcessState
		c.waitErr = err
		c.lock.Unlock()
		c.doneChan <- struct{}{}
	}()
	return nil
}

func (c *Cmd) Wait() error {
	<-c.doneChan
	return c.waitErr
}

func (c *Cmd) Running() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.processState == nil
}

func (c *Cmd) ExitCode() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.processState != nil {
		return c.processState.ExitCode()
	}
	return -1
}

// Shutdown will shut down the os process. If the provided context is cancelled
// the process will be killed with SIGKILL.
func (c *Cmd) Shutdown(ctx context.Context) error {
	{
		c.lock.RLock()
		if !c.started {
			c.lock.RUnlock()
			return fmt.Errorf("Cmd not running")
		}
		c.lock.RUnlock()
	}

	killSignal := os.Interrupt
	if err := c.Cmd.Process.Signal(os.Interrupt); err != nil {
		return err
	}
	var err error

	select {
	case <-c.doneChan:
		err = c.waitErr
	case <-ctx.Done():
		killSignal = os.Kill
		// Again, see above:
		// https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
		_ = syscall.Kill(-c.Cmd.Process.Pid, syscall.SIGKILL)
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
		select {
		case <-c.doneChan:
			err = c.waitErr
		case <-ctx.Done():
			// Although we are trying to SIGKILL all children by setting Setpgid
			// we cannot account for grandchildren that have been spawned with
			// their own process group id. This will only prevent .Wait from
			// being returned if os.Stdout/err are not a *os.File. See
			// https://github.com/golang/go/issues/23019
			err = fmt.Errorf("process did not exit")
		}
		cancel()
	}
	close(c.doneChan)
	if er, match := err.(*exec.ExitError); match {
		// If we are exiting because of the status code we expected to send
		if int(er.Sys().(syscall.WaitStatus)) == int(killSignal.(syscall.Signal)) {
			return nil
		}
	}
	return err
}
