package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gobuild/log"
	"github.com/howeyc/fsnotify"
)

type gowatch struct {
	Path           string        `json:"path"`
	Depth          int           `json:"depth"`
	Include        string        `json:"include"`
	Exclude        string        `json:"exclude"`
	BufferDuration string        `json:"buffer-duration"`
	bufdur         time.Duration `json:"-"`

	Command       []string          `json:"command"`
	Env           map[string]string `json:"env"`
	EnableRestart bool              `json:"enable-restart"`

	w       *fsnotify.Watcher `json:"-"`
	modtime map[string]time.Time
	sig     chan string
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
			base := info.Name()
			if base != "." && strings.HasPrefix(base, ".") { // ignore hidden dir
				return filepath.SkipDir
			}

			pathDepth := strings.Count(path, string(os.PathSeparator)) - baseNumSeps
			if pathDepth > this.Depth {
				return filepath.SkipDir
			}
			fmt.Println(">>> watch dir: ", path)
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
	this.modtime = make(map[string]time.Time)
	this.sig = make(chan string)
	go this.drainExec()
	this.drainEvent()
	return
}

func (this *gowatch) drainEvent() {
	go func() {
		for {
			log.Warnf("watch error: %s", <-this.w.Error)
		}
	}()
	for {
		eve := <-this.w.Event
		log.Debug(eve)
		changed := this.IsfileChanged(eve.Name)
		if changed {
			log.Info(eve)
			select {
			case this.sig <- "KILL":
			default:
			}
		}
	}
}

func (this *gowatch) IsfileChanged(p string) bool {
	p = filepath.Clean(p)
	fi, err := os.Stat(p)
	if err != nil {
		return true // if file not exists, just return true
	}
	curr := fi.ModTime()
	defer func() { this.modtime[p] = curr }()
	modt, ok := this.modtime[p]
	return !ok || curr.After(modt.Add(time.Second))
}

func (this *gowatch) drainExec() {
	for {
		startTime := time.Now()
		cmd := this.Command
		if len(cmd) == 0 {
			cmd = []string{"echo", "no command specified"}
		}
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stdout
		err := c.Start()
		if err != nil {
			log.Warn(err)
		}
		select {
		case <-this.sig:
			if c.Process != nil {
				c.Process.Kill()
			}
		case err = <-Go(c.Wait):
			if err != nil {
				log.Warn(err)
			}
		}
		if time.Since(startTime) > time.Second*3 {
			continue
		}
		<-this.sig
	}
}

func Go(f func() error) chan error {
	ch := make(chan error)
	go func() {
		ch <- f()
	}()
	return ch
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

const JSONCONF = ".fswatch.json"

func main() {
	notifyDelay, err := time.ParseDuration("1s")
	if err != nil {
		notifyDelay = time.Millisecond * 500
	}
	_ = notifyDelay
	gw := &gowatch{}
	gw.Path = "."
	gw.BufferDuration = "1s"
	gw.Depth = 0
	gw.Env = map[string]string{"PROGRAM": "fswatch"}
	gw.Command = []string{"echo", "hello fswatch!!!"}
	if fd, err := os.Open(JSONCONF); err == nil {
		if er := json.NewDecoder(fd).Decode(gw); er != nil {
			log.Fatal(er)
		}
		for key, val := range gw.Env {
			os.Setenv(key, val)
		}
		log.Fatal(gw.Watch())
	} else {
		fmt.Printf("initial %s file [y/n]: ", JSONCONF)
		var yn string = "y"
		fmt.Scan(&yn)
		if strings.ToUpper(strings.TrimSpace(yn)) == "Y" {
			data, _ := json.MarshalIndent(gw, "", "    ")
			ioutil.WriteFile(JSONCONF, data, 0644)
			fmt.Printf("%s created, use notepad++ or vim to edit it\n", strconv.Quote(JSONCONF))
		}
	}
}
