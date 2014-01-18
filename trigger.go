package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/shxsun/ansiterm"
	"github.com/shxsun/fswatch/termsize"
)

var runChannel = make(chan bool)
var screenLock = &sync.Mutex{}

func drainExec(name string, args ...string) {
	cmdFiled := &ansiterm.ScreenField{}
	var execTimes = 0
	for {
		<-runChannel

		screenLock.Lock()
		//ansiterm.SetBGColor(1)
		cmdFiled.SetRCW(sepLine+1, 0, termsize.Width()*(termsize.Height()-sepLine))
		ansiterm.MoveToXY(6, 0)
		cmdFiled.Erase()

		execTimes += 1
		prompt := fmt.Sprintf(" Start(%d) ", execTimes)
		termsize.Println(StringCenter(prompt, termsize.Width()))
		cmd := exec.Command(name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Println(err)
		}
		termsize.Println(StringCenter("  END  ", termsize.Width()))
		screenLock.Unlock()
	}
}
