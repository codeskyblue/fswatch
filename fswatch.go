package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gobuild/log"
	"github.com/howeyc/fsnotify"
)

var verbose = flag.Bool("v", false, "show verbose")

func init() {
	log.SetFlags(0)
	if runtime.GOOS == "windows" {
		log.SetPrefix("fswatch >>> ")
	} else {
		log.SetPrefix("\033[32mfswatch\033[0m >>> ")
	}
}

func colorPrintf(ansiColor string, format string, args ...interface{}) {
	if runtime.GOOS != "windows" {
		format = "\033[" + ansiColor + "m" + format + "\033[0m"
	}
	log.Printf(format, args...)
}

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

	AutoRestart     bool          `json:"autorestart"`
	RestartInterval time.Duration `json:"restart-interval"`
	KillSignal      string        `json:"kill-signal"`

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
		colorPrintf("35", "exec start")
		c := StartCmd(cmd[0], cmd[1:]...)
		err := c.Start()
		if err != nil {
			colorPrintf("35", err.Error())
		}
		select {
		case msg = <-this.sig:
			colorPrintf("33", "program terminated, signal(%s)", this.KillSignal)
			if err := KillCmd(c, this.KillSignal); err != nil {
				log.Errorf("group kill: %v", err)
			}
			if msg == "EXIT" {
				os.Exit(1)
			}
			goto SKIP_WAITING
		case err = <-Go(c.Wait):
			if err != nil {
				log.Warnf("program exited: %v", err)
			}
		}
		log.Infof("finish in %s", time.Since(startTime))
		if this.AutoRestart {
			goto SKIP_WAITING
		}
		colorPrintf("33", "-- wait signal --")
		if msg = <-this.sig; msg == "EXIT" {
			os.Exit(1)
		}
	SKIP_WAITING:
		if this.RestartInterval > 0 {
			log.Infof("restart after %s", this.RestartInterval)
		}
		time.Sleep(this.RestartInterval)
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
	gw := &gowatch{
		Paths:           []string{"."},
		Depth:           2,
		Exclude:         []string{},
		Include:         []string{"\\.(go|py|php|java|cpp|h|rb)$"},
		AutoRestart:     false,
		RestartInterval: 0,
		KillSignal:      "KILL",
	}
	gw.Env = map[string]string{"POWERD_BY": "github.com/codeskyblue/fswatch"}
	// load JSONCONF
	if fd, err := os.Open(JSONCONF); err == nil {
		if er := json.NewDecoder(fd).Decode(gw); er != nil {
			log.Fatal(er)
		}
		for key, val := range gw.Env {
			os.Setenv(key, val)
		}
	}
	flag.DurationVar(&gw.RestartInterval, "ri", gw.RestartInterval, "restart interval")
	flag.BoolVar(&gw.AutoRestart, "r", gw.AutoRestart, "enable autorestart")
	flag.StringVar(&gw.KillSignal, "k", gw.KillSignal, "kill signal")
	flag.Parse()
	if flag.NArg() > 0 {
		gw.Command = flag.Args()
	}
	if len(gw.Command) == 0 {
		fmt.Printf("initial %s file [y/n]: ", JSONCONF)
		var yn string = "y"
		fmt.Scan(&yn)
		gw.Command = []string{"bash", "-c", "whoami"}
		if strings.ToUpper(strings.TrimSpace(yn)) == "Y" {
			data, _ := json.MarshalIndent(gw, "", "    ")
			ioutil.WriteFile(JSONCONF, data, 0644)
			fmt.Printf("use notepad++ or vim to edit %s\n", strconv.Quote(JSONCONF))
		}
		return
	}
	log.Fatal(gw.Watch())
}
