package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/maxmcd/steady/internal/boxpool"
	"github.com/maxmcd/steady/internal/execx"
	"github.com/mitchellh/go-ps"
)

func removeTmpFiles() error {
	tmpFiles, err := filepath.Glob("/tmp/*")
	if err != nil {
		return err
	}
	for _, f := range tmpFiles {
		if err := os.RemoveAll(f); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	return nil
}

func main() {
	var cmd *execx.Cmd

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var action boxpool.ContainerAction
		lineBytes := scanner.Bytes()
		if err := json.Unmarshal(lineBytes, &action); err != nil {
			sendError(err)
			continue
		}
		switch action.Action {
		case "run":
			logs, err := os.Create("/opt/log.log")
			if err != nil {
				sendError(err)
				continue
			}
			exec := action.Exec
			cmd = execx.Command(exec.Cmd[0], exec.Cmd[1:]...)
			cmd.SysProcAttr = &syscall.SysProcAttr{}
			cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(exec.Uid), Gid: uint32(exec.Gid)}
			cmd.Env = exec.Env
			cmd.Dir = "/opt/app"
			cmd.Stdout = io.MultiWriter(logs, os.Stderr)
			cmd.Stderr = io.MultiWriter(logs, os.Stderr)
			if err := cmd.Start(); err != nil {
				cmd = nil
				sendError(err)
				continue
			}
			sendResponse(boxpool.ContainerResponse{Err: ""})
		case "stop":
			if cmd == nil {
				sendError(fmt.Errorf("no command running, cannot stop"))
				continue
			}
			if !cmd.Running() {
				sendResponse(boxpool.ContainerResponse{Err: ""})
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
			err := cmd.Shutdown(ctx)
			cancel()
			if err != nil && !strings.Contains(err.Error(), "process already finished") {
				sendError(fmt.Errorf("shutting down process: %w", err))
			}
			cmd = nil

			sendResponse(boxpool.ContainerResponse{Err: ""})
			// if err := removeTmpFiles(); err != nil {
			// 	fmt.Fprintln(os.Stderr, fmt.Errorf("removing editable files: %w", err))
			// }
			// Kill any orphaned processes
			processList, _ := ps.Processes()
			for _, p := range processList {
				if p.Pid() != 1 {
					_ = syscall.Kill(p.Pid(), syscall.SIGKILL)
				}
			}
		case "status":
			running := cmd.Running()
			exitCode := cmd.ExitCode()
			sendResponse(boxpool.ContainerResponse{
				Running:  &running,
				ExitCode: &exitCode,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func sendError(err error) {
	sendResponse(boxpool.ContainerResponse{Err: err.Error()})
}

func sendResponse(resp boxpool.ContainerResponse) {
	_ = json.NewEncoder(os.Stdout).Encode(resp)
}

func userToCred(u *user.User) *syscall.Credential {
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)
	return &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
}
