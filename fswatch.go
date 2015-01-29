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

const (
	CYELLOW = "33"
	CGREEN  = "32"
	CPURPLE = "35"
)

func cprintf(ansiColor string, format string, args ...interface{}) {
	if runtime.GOOS != "windows" {
		format = "\033[" + ansiColor + "m" + format + "\033[0m"
	}
	log.Printf(format, args...)
}

/*
type pathWatch struct {
	Include   string `json:"include"`
	reInclude *regexp.Regexp
	Exclude   string `json:"exclude"`
	reExclude *regexp.Regexp
	Depth     int `json:"depth"`
}

type fsWatch struct {
	PathWatches []pathWatch
	Command     []string `json:"command"`
	Cmd         string   `json:"cmd"` // if empty will add prefix(bash -c) and replace Command
	Signal      string   `json:"signal"`
	KillAsGroup bool     `json:"killasgroup"`
}

func init() {
	fw := fsWatch{
		PathWatches: []pathWatch{
			pathWatch{Include: "./", Exclude: "\\.svn"},
		},
		Cmd: "ls -l",
	}
	data, err := yaml.Marshal(fw)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}
*/

type gowatch struct {
	Paths     []string `json:"paths"`
	Depth     int      `json:"depth"`
	Exclude   []string `json:"exclude"`
	reExclude []*regexp.Regexp
	Include   []string `json:"include"`
	reInclude []*regexp.Regexp
	bufdur    time.Duration `json:"-"`
	Command   interface{}   `json:"command"` // can be string or []string
	cmd       []string
	Env       map[string]string `json:"env"`

	AutoRestart     bool          `json:"autorestart"`
	RestartInterval time.Duration `json:"restart-interval"`
	KillSignal      string        `json:"kill-signal"`

	w       *fsnotify.Watcher
	modtime map[string]time.Time
	sig     chan string
	sigOS   chan os.Signal
}

// Check if file matches
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
func (this *gowatch) watchDirAndChildren(path string, depth int) error {
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
			if pathDepth > depth {
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

// Create a fsnotify fswatch
// Initial vars
func (this *gowatch) Watch() (err error) {
	if this.w, err = fsnotify.NewWatcher(); err != nil {
		return
	}
	for _, path := range this.Paths {
		// translate env-vars
		if err = this.watchDirAndChildren(os.ExpandEnv(path), this.Depth); err != nil {
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

// filter fsevent and send to this.sig
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

// Use modified time to judge if file changed
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
	log.Println("command:", this.cmd)
	var msg string
	for {
		startTime := time.Now()
		cmd := this.cmd
		if len(cmd) == 0 {
			cmd = []string{"echo", "no command specified"}
		}
		cprintf(CGREEN, "exec start")
		c := StartCmd(cmd[0], cmd[1:]...)
		// Start to run command
		err := c.Start()
		if err != nil {
			cprintf("35", err.Error())
		}
		// Wait until killed or finished
		select {
		case msg = <-this.sig:
			cprintf(CYELLOW, "program terminated, signal(%s)", this.KillSignal)
			if err := KillCmd(c, this.KillSignal); err != nil {
				log.Errorf("group kill: %v", err)
			}
			if msg == "EXIT" {
				os.Exit(1)
			}
			goto SKIP_WAITING
		case err = <-Go(c.Wait):
			if err != nil {
				cprintf(CPURPLE, "program exited: %v", err)
			}
		}
		log.Infof("finish in %s", time.Since(startTime))

		// Whether to restart right now
		if this.AutoRestart {
			goto SKIP_WAITING
		}
		cprintf("33", "-- wait signal --")
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

const JSONCONF = ".fswatch.json"

var (
	gw = &gowatch{
		Paths:           []string{"."},
		Depth:           2,
		Exclude:         []string{},
		Include:         []string{"\\.(go|py|php|java|cpp|h|rb)$"},
		AutoRestart:     false,
		RestartInterval: 0,
		KillSignal:      "KILL",
	}
	confExists = false
	extInclude string
)

// parse command flag
func flagParse() {
	gw.Env = map[string]string{"POWERD_BY": "github.com/codeskyblue/fswatch"}
	// load JSONCONF
	if fd, err := os.Open(JSONCONF); err == nil {
		if er := json.NewDecoder(fd).Decode(gw); er != nil {
			log.Fatalf("json decode error: %v", er)
		}
		for key, val := range gw.Env {
			os.Setenv(key, val)
		}
		confExists = true
	}
	flag.DurationVar(&gw.RestartInterval, "ri", gw.RestartInterval, "restart interval")
	flag.BoolVar(&gw.AutoRestart, "r", gw.AutoRestart, "auto restart")
	flag.StringVar(&gw.KillSignal, "k", gw.KillSignal, "kill signal")
	flag.StringVar(&extInclude, "ext", "", "extention eg: [cpp,c,h]")
	flag.Parse()
}

func main() {
	flag.Usage = func() {
		fmt.Printf(`Usage: fswatch [OPTIONS...] <arg...>
OPTIONS:
    -ext=""    Extenton seprated by , eg: cpp,c,h
    -v         Show verbose info

Example:
    fswatch -ext cpp,c,h make -f Makefile
`)
	}
	flagParse()
	if len(os.Args) == 1 && !confExists {
		fmt.Printf("Create %s file [y/n]: ", strconv.Quote(JSONCONF))
		var yn string = "y"
		fmt.Scan(&yn)
		gw.Command = "echo helloworld"
		if strings.ToUpper(strings.TrimSpace(yn)) == "Y" {
			data, _ := json.MarshalIndent(gw, "", "    ")
			ioutil.WriteFile(JSONCONF, data, 0644)
			fmt.Printf("use notepad++ or vim to edit %s\n", strconv.Quote(JSONCONF))
		}
		return
	}

	if flag.NArg() > 0 {
		gw.Command = []string(flag.Args())
	}
	if extInclude != "" {
		for _, ext := range strings.Split(extInclude, ",") {
			gw.Include = append(gw.Include, "\\."+ext+"$")
		}
	}

	switch gw.Command.(type) {
	default:
		log.Fatal("check you config file. \"command\" must be string or []string")
	case string:
		if runtime.GOOS == "windows" {
			gw.cmd = []string{"cmd", "/c", gw.Command.(string)}
		} else {
			gw.cmd = []string{"bash", "-c", gw.Command.(string)}
		}
	case []string:
		gw.cmd = gw.Command.([]string)
	}

	log.Fatal(gw.Watch())
}
