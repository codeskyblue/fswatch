package main

import (
	"os"
	"os/exec"

	"github.com/hotei/ansiterm"
	"github.com/shxsun/fswatch/termsize"
)

var runChannel = make(chan bool)

func drainExec(name string, args ...string) {
	for {
		<-runChannel

		ansiterm.ClearPage()

		termsize.Println(StringCenter(" START ", TermSize))
		cmd := exec.Command(name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		err := cmd.Run()
		if err != nil {
			logs.Error(err)
		}
		termsize.Println(StringCenter("  END  ", TermSize))
	}
}
