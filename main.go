package main

import (
	"flag"
	"fmt"
	"github.com/howeyc/fsnotify"
	"github.com/shxsun/klog"
	"os"
	_ "os/exec"
)

var (
	eventTime = make(map[string]int64)
	K         = klog.NewLogger(nil, "")
)

func watchEvent(watcher *fsnotify.Watcher) {
	for {
		select {
		case ev := <-watcher.Event:
			K.Info(ev)
			/*
				isbuild := true

				// Skip TMP files for Sublime Text.
				if checkTMPFile(e.Name) {
					continue
				}
				if !checkIsGoFile(e.Name) {
					continue
				}

				mt := getFileModTime(e.Name)
				if t := eventTime[e.Name]; mt == t {
					K.Debugf("skip %s", e.String())
					isbuild = false
				}

				eventTime[e.Name] = mt

				if isbuild {
					K.Info(e)
					// restart program
					go Autobuild()
				}
			*/
		case err := <-watcher.Error:
			K.Warn("watch error: %s", err) // No need to exit here
		}
	}
}

func NewWatcher(paths []string) {
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
	watchEvent(w)
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Printf("Usage: %s -a [watch_path] cmd [args...]\n", os.Args[0])
		return
	}
	NewWatcher([]string{"."})
}
