package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/maxmcd/steady/daemon/boxpool"
)

func main() {
	// user, err := user.Current()
	// if err != nil {
	// 	panic(err)
	// }

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var action boxpool.ContainerAction
		lineBytes := scanner.Bytes()
		if err := json.Unmarshal(lineBytes, &action); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		fmt.Fprintln(os.Stderr, "got action", action)
		if action.Action == "run" {
			logs, err := os.Create("/opt/log.log")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			cmd := exec.Command(action.Cmd[0], action.Cmd[1:]...)
			cmd.Env = action.Env
			cmd.Dir = "/opt/app"
			cmd.Stdout = logs
			cmd.Stderr = logs
			if err := cmd.Start(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
