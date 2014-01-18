package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/howeyc/fsnotify"
	"github.com/jessevdk/go-flags"
	"github.com/shxsun/ansiterm"
	"github.com/shxsun/fswatch/termsize"
	"github.com/shxsun/klog"
)

var (
	reader, writer = io.Pipe()
	screen         = NewBufReader(reader)
	logs           = klog.NewLogger(writer, "").
			SetLevel(klog.LInfo).SetFlags(klog.Fdevflag)
	notifyDelay time.Duration
	LangExts    []string
	headHeight  = 4
	sepLine     = 5
)

func goDrainScreen() {
	toplines := make([]string, headHeight)
	cur := -1
	field := new(ansiterm.ScreenField)
	field.SetRCW(2, 0, termsize.Width()*headHeight)
	scrbuf := bufio.NewReader(screen)
	go func() {
		for {
			line, err := scrbuf.ReadString('\n')
			if err != nil && err != io.EOF {
				log.Println(err)
				return
			}
			screenLock.Lock()
			cur = (cur + 1) % headHeight
			toplines[cur] = line
			field.Erase()
			for i := 0; i < headHeight; i++ {
				index := (cur + i) % headHeight
				fmt.Print(toplines[index])
			}
			screenLock.Unlock()
		}
	}()
}

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

func delayEvent(event chan *fsnotify.FileEvent, notifyDelay time.Duration) {
	for {
		select {
		case <-event: //filterEvent:
			continue
		case <-time.After(notifyDelay):
			return
		}
	}
}

func Watch(e chan *fsnotify.FileEvent) {
	filterEvent := filter(e, // filter useless event
		hiddenFilter,
		extentionFilter,
		gitignoreFilter,
		deleteRenameFilter,
		md5CheckFileter)

	for {
		ev := <-filterEvent
		logs.Info("Sense first: ", ev)
		delayEvent(filterEvent, notifyDelay)

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

	ansiterm.ClearPage()
	go drainExec(args[0], args[1:]...)
	goDrainScreen()

	if len(opts.Paths) == 0 {
		opts.Paths = append(opts.Paths, ".")
	}
	fmt.Println("fswatch:", opts.Paths, "--", args)

	LangExts = strings.Split(opts.Exts, ",")
	logs.Info("Watch extentions:", LangExts)
	Watch(NewEvent(opts.Paths, opts.Depth))
}
