package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gobuild/log"
	"github.com/howeyc/fsnotify"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("\033[32mfswatch\033[0m >>> ")
}

var (
	verbose = flag.Bool("v", false, "show verbose")
)

type gowatch struct {
	Paths     []string `json:"paths"`
	Depth     int      `json:"depth"`
	Exclude   []string `json:"exclude"`
	reExclude []*regexp.Regexp
	Include   []string `json:"include"`
	reInclude []*regexp.Regexp
	bufdur    time.Duration     `json:"-"`
	Command   []string          `json:"command"`
	Env       map[string]string `json:"env"`
	//EnableRestart bool              `json:"enable-restart"`

	w       *fsnotify.Watcher
	modtime map[string]time.Time
	sig     chan string
	sigOS   chan os.Signal
}

func (this *gowatch) match(file string) bool {
	file = filepath.Base(file)
	for _, rule := range this.reExclude {
		if rule.MatchString(file) {
			return false
		}
	}
	for _, rule := range this.reInclude {
		if rule.MatchString(file) {
			return true
		}
	}
	return len(this.reInclude) == 0 // if empty include, then return true
}

// Add dir and children (recursively) to watcher
func (this *gowatch) watchDirAndChildren(path string) error {
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
			if *verbose {
				fmt.Println(">>> watch dir: ", path)
			}
			if err := this.w.Watch(path); err != nil {
				return err
			}
		}
		return nil
	})
}

func (this *gowatch) Watch() (err error) {
	if this.w, err = fsnotify.NewWatcher(); err != nil {
		return
	}
	for _, path := range this.Paths {
		if err = this.watchDirAndChildren(os.ExpandEnv(path)); err != nil {
			log.Fatal(err)
		}
	}
	this.modtime = make(map[string]time.Time)
	this.sig = make(chan string)
	for _, patten := range this.Exclude {
		this.reExclude = append(this.reExclude, regexp.MustCompile(patten))
	}
	for _, patten := range this.Include {
		this.reInclude = append(this.reInclude, regexp.MustCompile(patten))
	}

	this.sigOS = make(chan os.Signal, 1)
	signal.Notify(this.sigOS, syscall.SIGINT)

	go this.drainExec()
	this.drainEvent()
	return
}

func (this *gowatch) drainEvent() {
	for {
		select {
		case err := <-this.w.Error:
			log.Warnf("watch error: %s", err)
		case <-this.sigOS:
			this.sig <- "EXIT"
		case eve := <-this.w.Event:
			log.Debug(eve)
			changed := this.IsfileChanged(eve.Name)
			if changed && this.match(eve.Name) {
				log.Info(eve)
				select {
				case this.sig <- "KILL":
				default:
				}
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
	log.Println("command:", this.Command)
	var msg string
	for {
		startTime := time.Now()
		cmd := this.Command
		if len(cmd) == 0 {
			cmd = []string{"echo", "no command specified"}
		}
		log.Info("\033[35mexec start\033[0m")
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stdout

		c.SysProcAttr = &syscall.SysProcAttr{}
		c.SysProcAttr.Setpgid = true

		err := c.Start()
		if err != nil {
			log.Warn(err)
		}
		select {
		case msg = <-this.sig:
			if err := groupKill(c); err != nil {
				log.Error(err)
			}
			if msg == "EXIT" {
				os.Exit(1)
			}
			goto SKIP_WAITING
		case err = <-Go(c.Wait):
			if err != nil {
				log.Warn(err)
			}
		}
		//if this.EnableRestart && time.Since(startTime) > time.Second*2 {
		//continue
		//}
		log.Infof("finish in %s", time.Since(startTime))
		log.Info("\033[33m-- wait signal --\033[0m")
		if msg = <-this.sig; msg == "EXIT" {
			os.Exit(1)
		}
	SKIP_WAITING:
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
	flag.Parse()
	gw := &gowatch{
		Paths:   []string{"."},
		Depth:   2,
		Command: []string{"echo", "fswatch"},
		Exclude: []string{},
		Include: []string{"\\.(go|py|php|java|cpp|h|rb)$"},
	}
	gw.Env = map[string]string{"POWERD_BY": "github.com/codeskyblue/fswatch"}
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
			fmt.Printf("use notepad++ or vim to edit %s\n", strconv.Quote(JSONCONF))
		}
	}
}
