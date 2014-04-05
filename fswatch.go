package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gobuild/log"
	"github.com/howeyc/fsnotify"
	"github.com/jessevdk/go-flags"
)

type gowatch struct {
	Path           string        `json:"path"`
	Depth          int           `json:"depth"`
	Include        string        `json:"include"`
	Exclude        string        `json:"exclude"`
	BufferDuration string        `json:"buffer-duration"`
	bufdur         time.Duration `json:"-"`

	Command       string   `json:"command"`
	Env           []string `json:"env"`
	EnableRestart bool     `json:"enable-restart"`

	w *fsnotify.Watcher `json:"-"`
}

// Add dir and children (recursively) to watcher
func (this *gowatch) watchDirAndChildren() error {
	path := this.Path
	if err := this.w.Watch(path); err != nil {
		return err
	}
	baseNumSeps := strings.Count(path, string(os.PathSeparator))
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if strings.HasPrefix(".", info.Name()) { // ignore hidden dir
				return filepath.SkipDir
			}

			pathDepth := strings.Count(path, string(os.PathSeparator)) - baseNumSeps
			if pathDepth > this.Depth {
				return filepath.SkipDir
			}
			if err := this.w.Watch(path); err != nil {
				return err
			}
		}
		return nil
	})
}

func (this *gowatch) Watch() (err error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	this.w = w
	err = this.watchDirAndChildren()
	if err != nil {
		return
	}
	go this.drainExec()
	return
}

func (this *gowatch) drainWarn() {
	for {
		err := <-this.w.Error
		log.Warnf("watch error: %s", err) // No need to exit here
	}
}

func (this *gowatch) drainExec() {
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

var opts struct {
	Verbose bool     `json:"verbose" short:"v" long:"verbose" description:"Show verbose debug infomation"`
	Delay   string   `json:"timegap" long:"delay" description:"Trigger event buffer time" default:"0.1s"`
	Depth   int      `json:"depth" short:"d" long:"depth" description:"depth of watch" default:"3"`
	Paths   []string `json:"paths" short:"p" long:"path" description:"watch path, support multi -p"`
}

var (
	args []string
)

var runChannel = make(chan bool)

func drainExec(name string, args ...string) {
	for {
		<-runChannel

		cmd := exec.Command(name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	parser := flags.NewParser(&opts, flags.Default|flags.PassAfterNonOption)
	parser.Usage = "fswatch [OPTION] command [args...]"
	var err error
	args, err = parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	notifyDelay, err := time.ParseDuration(opts.Delay)
	if err != nil {
		notifyDelay = time.Millisecond * 500
	}
	_ = notifyDelay
}
