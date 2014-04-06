// utils.go
package main

import (
	"fmt"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"

	"github.com/gobuild/log"
	"github.com/shxsun/go-sh"
)

func groupKill(cmd *exec.Cmd) (err error) {
	log.Println("\033[33mprogram terminated\033[0m")
	var pid, pgid int
	if cmd.Process != nil {
		pid = cmd.Process.Pid
		c := sh.Command("/bin/ps", "-o", "pgid", "-p", strconv.Itoa(pid)).Command("sed", "-n", "2,$p")
		var out []byte
		out, err = c.Output()
		if err != nil {
			return
		}
		_, err = fmt.Sscanf(string(out), "%d", &pgid)
		if err != nil {
			return
		}
		err = exec.Command("/bin/kill", "-TERM", "-"+strconv.Itoa(pgid)).Run()
	}
	return
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
