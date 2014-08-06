// utils.go
package main

import (
	"fmt"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"

	sh "github.com/codeskyblue/go-sh"
	"github.com/gobuild/log"
)

func groupKill(cmd *exec.Cmd, signal string) (err error) {
	log.Println("\033[33mprogram terminated\033[0m")
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

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
