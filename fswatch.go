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
	lg             = klog.NewLogger(writer, "").
			SetLevel(klog.LInfo).SetFlags(klog.Fstdflag)
	notifyDelay time.Duration
	LangExts    []string
	headHeight  = 4
	sepLine     = 5
)

func goDrainScreen() {
	toplines := make([]string, headHeight)
	cur := -1
	field := new(ansiterm.ScreenField)
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

			if !opts.Verbose {
				ansiterm.MoveToXY(0, 0)
				ansiterm.ClearLine()
			}
			fmt.Println("fswatch:", opts.Paths, "--", args)

			if !opts.Verbose {
				field.SetRCW(2, 0, termsize.Width()*headHeight)
			}
			if !opts.Verbose {
				field.Erase()
			}
			for i := 0; i < headHeight; i++ {
				index := (cur - i + headHeight) % headHeight
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
			if strings.HasPrefix(".", info.Name()) { // ignore hidden dir
				return filepath.SkipDir
			}

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
		log.Fatalf("fail to create new Watcher: %s", err)
	}
	for _, path := range paths {
		lg.Debugf("watch directory: %s", path)
		err = watchDirAndChildren(watcher, path, depth)
		if err != nil {
			lg.Error(err)
		}
	}

	// ignore watcher error
	go func() {
		for {
			err := <-watcher.Error           // ignore watcher error
			lg.Warnf("watch error: %s", err) // No need to exit here
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
		lg.Info("Sense first: ", ev)
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

var (
	args []string
)

func main() {
	parser := flags.NewParser(&opts, flags.Default|flags.PassAfterNonOption)
	parser.Usage = "fswatch [OPTION] command [args...]"
	var err error
	args, err = parser.Parse()
	if err != nil {
		os.Exit(1)
	}
	if opts.Verbose {
		lg.SetLevel(klog.LDebug)
	}
	notifyDelay, err = time.ParseDuration(opts.Delay)
	if err != nil {
		lg.Warn(err)
		notifyDelay = time.Millisecond * 500
	}
	lg.Debugf("delay time: %s", notifyDelay)

	if len(args) == 0 {
		fmt.Printf("Use %s --help for more details\n", os.Args[0])
		return
	}

	// check if cmd exists
	_, err = exec.LookPath(args[0])
	if err != nil {
		lg.Fatal(err)
	}

	if !opts.Verbose {
		ansiterm.ClearPage()
	}
	go drainExec(args[0], args[1:]...)
	goDrainScreen()

	if len(opts.Paths) == 0 {
		opts.Paths = append(opts.Paths, ".")
	}

	LangExts = strings.Split(opts.Exts, ",")
	lg.Info("Watch extentions:", LangExts)
	event := NewEvent(opts.Paths, opts.Depth)
	Watch(event)
}
