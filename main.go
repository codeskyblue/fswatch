package main

import (
	"flag"
	"fmt"
	"github.com/howeyc/fsnotify"
	"github.com/shxsun/klog"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	K = klog.NewLogger(nil, "")
)

func filter(watch chan *fsnotify.FileEvent) chan *fsnotify.FileEvent {
	modifyTime := make(map[string]int64)
	filterd := make(chan *fsnotify.FileEvent)
	go func() {
		for {
			ev := <-watch
			K.Debug("FILTER:", ev)
			if ev.IsDelete() || ev.IsRename() {
				filterd <- ev
				continue
			}
			mt, err := getFileModTime(ev.Name)
			if err != nil {
				//K.Warnf("get file mod time failed: %s", err)
				continue
			}
			if t := modifyTime[ev.Name]; mt != t {
				filterd <- ev
				modifyTime[ev.Name] = mt
			} else {
				K.Debugf("SKIP: %s", ev.Name)
			}
		}
	}()
	return filterd
}

// ths main goroutine
func watchEvent(watcher *fsnotify.Watcher, name string, args ...string) {
	go func() {
		for {
			err := <-watcher.Error          // ignore watcher error
			K.Warnf("watch error: %s", err) // No need to exit here
		}
	}()

	var cmd *exec.Cmd
	filterEvent := filter(watcher.Event)
	for {
		select {
		case ev := <-filterEvent:
			K.Info(ev)
		case <-time.After(notifyDelay):
			continue
		}
		// stop cmd
		if cmd != nil && cmd.Process != nil {
			K.Info("stop process")
			cmd.Process.Kill()
		}
		// start cmd
		cmd = exec.Command(name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Println(fmt.Sprintf("%s %5s %s", LeftRight, "START", LeftRight))
		err := cmd.Start()
		if err != nil {
			K.Error(err)
			continue
		} else {
			go func(cmd *exec.Cmd) {
				cmd.Wait()
				fmt.Println(fmt.Sprintf("%s %5s %s", LeftRight, "E N D", LeftRight))
			}(cmd)
		}
	}
}

func NewWatcher(paths []string, name string, args ...string) {
	//K.SetLevel(klog.LDebug)
	w, err := fsnotify.NewWatcher()
	if err != nil {
		K.Fatalf("fail to create new Watcher: %s", err)
	}

	K.Info("initial watcher")
	for _, path := range paths {
		K.Debug("watch directory: %s", path)
		err = w.Watch(path)
		if err != nil {
			K.Fatal("fail to watch directory: %s", err)
		}
	}
	watchEvent(w, name, args...)
}

var (
	notifyDelay time.Duration = time.Second * 1
	LeftRight                 = strings.Repeat("-", 25)
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Printf("Usage: %s -a [watch_path] cmd [args...]\n", os.Args[0])
		return
	}
	cmd := exec.Command(flag.Arg(0))
	cmd.Args = flag.Args()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	NewWatcher([]string{"."}, flag.Arg(0), flag.Args()[1:]...)
}
