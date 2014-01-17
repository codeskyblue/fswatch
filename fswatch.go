package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/howeyc/fsnotify"
	"github.com/jessevdk/go-flags"
	"github.com/shxsun/fswatch/termsize"
	"github.com/shxsun/klog"
)

var (
	logs        = klog.DevLog.SetLevel(klog.LInfo)
	notifyDelay time.Duration
	LangExts    []string
)

// Add dir and children (recursively) to watcher
func watchDirAndChildren(w *fsnotify.Watcher, path string, depth int) error {
	if err := w.Watch(path); err != nil {
		return err
	}
	baseNumSeps := strings.Count(path, string(os.PathSeparator))
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			pathDepth := strings.Count(path, string(os.PathSeparator)) - baseNumSeps
			if pathDepth > depth {
				return filepath.SkipDir
			}
			if opts.Verbose {
				fmt.Fprintln(os.Stderr, "Watching", path)
			}
			if err := w.Watch(path); err != nil {
				return err
			}
		}
		return nil
	})
}

// generate new event
func NewEvent(paths []string, depth int) chan *fsnotify.FileEvent {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logs.Fatalf("fail to create new Watcher: %s", err)
	}
	logs.Info("Initial watcher")

	for _, path := range paths {
		logs.Debugf("watch directory: %s", path)
		err = watchDirAndChildren(watcher, path, depth)
		if err != nil {
			logs.Fatalf("failed to watch directory: %s", err)
		}
	}

	// ignore watcher error
	go func() {
		for {
			err := <-watcher.Error             // ignore watcher error
			logs.Warnf("watch error: %s", err) // No need to exit here
		}
	}()

	return watcher.Event
}

var running = false
var runChannel = make(chan bool)

func drainExec(name string, args ...string) {
	for {
		<-runChannel
		termsize.Println(StringCenter(" START ", TermSize))
		cmd := exec.Command(name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		err := cmd.Run()
		if err != nil {
			logs.Error(err)
		}
		termsize.Println(StringCenter(" END ", TermSize))
	}
}

func Watch(e chan *fsnotify.FileEvent) {
	// filter useless event
	filterEvent := filter(e,
		hiddenFilter,
		extentionFilter,
		gitignoreFilter,
		deleteRenameFilter,
		md5CheckFileter)

	for {
		ev := <-filterEvent
		logs.Info("Sense first: ", ev)
	CHECK:
		select {
		case ev = <-filterEvent:
			logs.Info("Sense again: ", ev)
			goto CHECK
		case <-time.After(notifyDelay):
		}
		select {
		case runChannel <- true:
		default:
		}
	}
}

var opts struct {
	Verbose bool     `short:"v" long:"verbose" description:"Show verbose debug infomation"`
	Delay   string   `long:"delay" description:"Trigger event buffer time" default:"0.1s"`
	Depth   int      `short:"d" long:"depth" description:"depth of watch" default:"3"`
	Exts    string   `short:"e" long:"ext" description:"only watch specfied ext file" default:"go,py,c,rb,cpp,cxx,h"`
	Paths   []string `short:"p" long:"path" description:"watch path, support multi -p"`
}

func main() {
	parser := flags.NewParser(&opts, flags.Default|flags.PassAfterNonOption)
	args, err := parser.Parse()

	if err != nil {
		os.Exit(1)
	}
	if opts.Verbose {
		logs.SetLevel(klog.LDebug)
	}
	notifyDelay, err = time.ParseDuration(opts.Delay)
	if err != nil {
		logs.Warn(err)
		notifyDelay = time.Millisecond * 500
	}
	logs.Debugf("delay time: %s", notifyDelay)

	if len(args) == 0 {
		fmt.Printf("Use %s --help for more details\n", os.Args[0])
		return
	}

	// check if cmd exists
	_, err = exec.LookPath(args[0])
	if err != nil {
		logs.Fatal(err)
	}
	go drainExec(args[0], args[1:]...)

	if len(opts.Paths) == 0 {
		opts.Paths = append(opts.Paths, ".")
	}
	logs.Info(opts.Paths)

	LangExts = strings.Split(opts.Exts, ",")
	logs.Info(LangExts)
	Watch(NewEvent(opts.Paths, opts.Depth))
}
