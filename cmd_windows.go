package main

import (
	"os"
	"os/exec"
	"strconv"
)

func StartCmd(name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stdout
	return c
}

func KillCmd(cmd *exec.Cmd, signal string) (err error) {
	var pid int
	if cmd.Process != nil {
		pid = cmd.Process.Pid
		c := exec.Command("taskkill", "/t", "/f", "/pid", strconv.Itoa(pid))
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()
	}
	return
}
