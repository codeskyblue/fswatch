package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/codeskyblue/go-sh"
)

func StartCmd(name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stdout

	c.SysProcAttr = &syscall.SysProcAttr{}
	c.SysProcAttr.Setpgid = true
	return c
}

func KillCmd(cmd *exec.Cmd, signal string) (err error) {
	var pid, pgid int
	if cmd.Process != nil {
		pid = cmd.Process.Pid
		sess := sh.NewSession()
		if *verbose {
			sess.ShowCMD = true
		}
		c := sess.Command("/bin/ps", "-o", "pgid", "-p", strconv.Itoa(pid)).Command("sed", "-n", "2,$p")
		var out []byte
		out, err = c.Output()
		if err != nil {
			return
		}
		_, err = fmt.Sscanf(string(out), "%d", &pgid)
		if err != nil {
			return
		}
		err = sess.Command("pkill", "-"+signal, "--pgroup", strconv.Itoa(pgid)).Run()
	}
	return
}
