package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	ignore "github.com/codeskyblue/dockerignore"
	"github.com/codeskyblue/kexec"
	"github.com/go-fsnotify/fsnotify"
	"github.com/gobuild/log"
	"github.com/google/shlex"
	yaml "gopkg.in/yaml.v2"
)

const (
	FWCONFIG_YAML = ".fsw.yml"
	FWCONFIG_JSON = ".fsw.json"
)

var (
	VERSION = "2.3"
)

var signalMaps = map[string]os.Signal{
	"INT":  syscall.SIGINT,
	"HUP":  syscall.SIGHUP,
	"QUIT": syscall.SIGQUIT,
	"TRAP": syscall.SIGTRAP,
	"TERM": syscall.SIGTERM,
	"KILL": syscall.SIGKILL, // kill -9
}

func init() {
	for key, val := range signalMaps {
		signalMaps["SIG"+key] = val
		signalMaps[fmt.Sprintf("%d", val)] = val
	}
	log.SetFlags(0)
	if runtime.GOOS == "windows" {
		log.SetPrefix("fswatch >>> ")
	} else {
		log.SetPrefix("\033[32mfswatch\033[0m >>> ")
	}
}

const (
	CBLACK   = "30"
	CRED     = "31"
	CGREEN   = "32"
	CYELLOW  = "33"
	CBLUE    = "34"
	CMAGENTA = "35"
	CPURPLE  = "36"
)

func CPrintf(ansiColor string, format string, args ...interface{}) {
	if runtime.GOOS != "windows" {
		format = "\033[" + ansiColor + "m" + format + "\033[0m"
	}
	log.Printf(format, args...)
}

type FSEvent struct {
	Name string
}

type FWConfig struct {
	Description string         `yaml:"desc" json:"desc"`
	Triggers    []TriggerEvent `yaml:"triggers" json:"triggers"`
	WatchPaths  []string       `yaml:"watch_paths" json:"watch_paths"`
	WatchDepth  int            `yaml:"watch_depth" json:"watch_depth"`
}

type TriggerEvent struct {
	Name                string            `yaml:"name" json:"name"`
	Patterns            []string          `yaml:"patterns" json:"patterns"`
	matchPatterns       []string          `yaml:"-" json:"-"`
	Environ             map[string]string `yaml:"env" json:"env"`
	Command             string            `yaml:"cmd" json:"cmd"`
	Shell               bool              `yaml:"shell" json:"shell"`
	cmdArgs             []string          `yaml:"-" json:"-"`
	Delay               string            `yaml:"delay" json:"delay"`
	delayDuration       time.Duration     `yaml:"-" json:"-"`
	StopTimeout         string            `yaml:"stop_timeout" json:"stop_timeout"`
	stopTimeoutDuration time.Duration     `yaml:"-" json:"-"`
	Signal              string            `yaml:"signal" json:"signal"`
	killSignal          os.Signal         `yaml:"-" json:"-"`
	KillSignal          string            `yaml:"kill_signal" json:"kill_signal"`
	exitSignal          os.Signal
	kcmd                *kexec.KCommand
}

func (this *TriggerEvent) Start() (waitC chan error) {
	CPrintf(CGREEN, fmt.Sprintf("[%s] exec start: %v", this.Name, this.cmdArgs))
	startTime := time.Now()
	cmd := kexec.Command(this.cmdArgs[0], this.cmdArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	env := os.Environ()
	for key, val := range this.Environ {
		env = append(env, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Env = env
	this.kcmd = cmd
	waitC = make(chan error, 1)
	if err := cmd.Start(); err != nil {
		waitC <- err
		return
	}
	go func() {
		//var er error
		waitC <- cmd.Wait()
		log.Infof("[%s] finish in %s", this.Name, time.Since(startTime))
	}()
	return waitC
}

func (this *TriggerEvent) Stop(waitC chan error) bool {
	if this.kcmd != nil {
		if this.kcmd.ProcessState != nil && this.kcmd.ProcessState.Exited() {
			this.kcmd = nil
			return true
		}
		this.kcmd.Terminate(this.killSignal)
		var done bool
		select {
		case err := <-waitC:
			if err != nil {
				CPrintf(CRED, "[%s] program exited: %v", this.Name, err)
			}
			done = true
		case <-time.After(this.stopTimeoutDuration):
			done = false
		}
		if !done {
			CPrintf(CYELLOW, "[%s] program still alive", this.Name)
			//<-waitC
			//CPrintf(CYELLOW, "[%s] program still alive, force kill", this.Name)
			//	this.kcmd.Terminate(syscall.SIGKILL)
		} else {
			this.kcmd = nil
		}
		return done
	}
	return true
}

// when use func (this *TriggerEvent) strange things happened, wired
func (this *TriggerEvent) WatchEvent(evtC chan FSEvent, wg *sync.WaitGroup) {
	waitC := this.Start()
	for evt := range evtC {
		isMatch, err := ignore.Matches(evt.Name, this.Patterns)
		if err != nil {
			log.Fatal(err)
		}
		if !isMatch {
			continue
		}
		if this.Stop(waitC) {
			CPrintf(CGREEN, "changed: %v", evt.Name)
			CPrintf(CGREEN, "delay: %v", this.Delay)
			time.Sleep(this.delayDuration)
			waitC = this.Start()
		}
	}

	// force kill when exit
	this.killSignal = this.exitSignal
	this.Stop(waitC)
	wg.Done()
}

func getShell() ([]string, error) {
	if path, err := exec.LookPath("bash"); err == nil {
		return []string{path, "-c"}, nil
	}
	if path, err := exec.LookPath("sh"); err == nil {
		return []string{path, "-c"}, nil
	}
	// even windows, there still has git-bash or mingw
	if runtime.GOOS == "windows" {
		return []string{"cmd", "/c"}, nil
	}
	return nil, fmt.Errorf("Could not find bash or sh on path.")
}

func fixFWConfig(in FWConfig) (out FWConfig, err error) {
	out = in
	for idx, trigger := range in.Triggers {
		outTg := &out.Triggers[idx]
		if trigger.Delay == "" {
			outTg.Delay = "100ms"
		}
		outTg.delayDuration, err = time.ParseDuration(outTg.Delay)
		if err != nil {
			return
		}
		if trigger.StopTimeout == "" {
			outTg.StopTimeout = "500ms"
		}
		outTg.stopTimeoutDuration, err = time.ParseDuration(outTg.StopTimeout)
		if err != nil {
			return
		}
		if outTg.Signal == "" {
			outTg.Signal = "KILL"
		}
		outTg.killSignal = signalMaps[outTg.Signal]
		if outTg.KillSignal == "" {
			outTg.exitSignal = syscall.SIGKILL
		} else {
			outTg.exitSignal = signalMaps[outTg.KillSignal]
		}

		rd := ioutil.NopCloser(bytes.NewBufferString(strings.Join(outTg.Patterns, "\n")))
		patterns, er := ignore.ReadIgnore(rd)
		if er != nil {
			err = er
			return
		}
		outTg.matchPatterns = patterns
		if outTg.Shell {
			sh, er := getShell()
			if er != nil {
				err = er
				return
			}
			outTg.cmdArgs = append(sh, outTg.Command)
		} else {
			outTg.cmdArgs, err = shlex.Split(outTg.Command)
			if err != nil {
				return
			}
			if len(outTg.cmdArgs) == 0 {
				err = errors.New("No command defined")
				return
			}
		}
	}
	if len(out.WatchPaths) == 0 {
		out.WatchPaths = append(out.WatchPaths, ".")
	}
	if out.WatchDepth < 0 {
		out.WatchDepth = 0
	}

	return
}

func readString(prompt, value string) string {
	fmt.Printf("[?] %s (%s) ", prompt, value)
	var s = value
	fmt.Scanf("%s", &s)
	return s
}

func genFWConfig() FWConfig {
	var (
		name    string
		command string
	)
	cwd, _ := os.Getwd()
	name = filepath.Base(cwd)
	name = readString("name:", name)

	for command == "" {
		fmt.Print("[?] command (go test -v): ")
		reader := bufio.NewReader(os.Stdin)
		command, _ = reader.ReadString('\n')
		command = strings.TrimSpace(command)
		if command == "" {
			command = "go test -v"
		}
	}
	fwc := FWConfig{
		Description: fmt.Sprintf("Auto generated by fswatch [%s]", name),
		Triggers: []TriggerEvent{{
			Patterns: []string{"**/*.go", "**/*.c", "**/*.py"},
			Environ: map[string]string{
				"DEBUG": "1",
			},
			Shell:   true,
			Command: command,
		}},
	}
	out, _ := fixFWConfig(fwc)
	return out
}

func ListAllDir(path string, depth int) (dirs []string, err error) {
	baseNumSeps := strings.Count(path, string(os.PathSeparator))
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			base := info.Name()
			if base != "." && strings.HasPrefix(base, ".") { // ignore hidden dir
				return filepath.SkipDir
			}
			if base == "node_modules" {
				return filepath.SkipDir
			}

			pathDepth := strings.Count(path, string(os.PathSeparator)) - baseNumSeps
			if pathDepth > depth {
				return filepath.SkipDir
			}
			dirs = append(dirs, path)
		}
		return nil
	})
	return
}

func UniqStrings(ss []string) []string {
	out := make([]string, 0, len(ss))
	m := make(map[string]bool, len(ss))
	for _, key := range ss {
		if !m[key] {
			out = append(out, key)
			m[key] = true
		}
	}
	return out
}

func IsDirectory(path string) bool {
	pinfo, err := os.Stat(path)
	return err == nil && pinfo.IsDir()
}

var fileModifyTimeMap = make(map[string]time.Time)

func IsChanged(path string) bool {
	pinfo, err := os.Stat(path)
	if err != nil {
		return true
	}
	mtime := pinfo.ModTime()
	if mtime.Sub(fileModifyTimeMap[path]) > time.Millisecond*100 { // 100ms
		fileModifyTimeMap[path] = mtime
		return true
	}
	return false
}

// visits here for in case of duplicate paths
func WatchPathAndChildren(w *fsnotify.Watcher, paths []string, depth int, visits map[string]bool) error {
	if visits == nil {
		visits = make(map[string]bool)
	}

	watchDir := func(dir string) error {
		if visits[dir] {
			return nil
		}
		if err := w.Add(dir); err != nil {
			if strings.Contains(err.Error(), "too many open files") {
				log.Fatalf("Watch directory(%s) error: %v", dir, err)
			}
			log.Warnf("Watch directory(%s) error: %v", dir, err)
			return err
		}
		log.Debug("Watch directory:", dir)
		//log.Info("Watch directory:", dir)
		visits[dir] = true
		return nil
	}
	var err error
	for _, path := range paths {
		if visits[path] {
			continue
		}

		watchDir(path)
		dirs, er := ListAllDir(path, depth)
		if er != nil {
			err = er
			log.Warnf("ERR list dir: %s, depth: %d, %v", path, depth, err)
			continue
		}

		for _, dir := range dirs {
			watchDir(dir)
		}
	}
	return err
}

func drainEvent(fwc FWConfig) (globalEventC chan FSEvent, wg *sync.WaitGroup, err error) {
	globalEventC = make(chan FSEvent, 1)
	wg = &sync.WaitGroup{}
	evtChannls := make([]chan FSEvent, 0)
	// log.Println(len(fwc.Triggers))
	for _, tg := range fwc.Triggers {
		wg.Add(1)
		evtC := make(chan FSEvent, 1)
		evtChannls = append(evtChannls, evtC)
		go func(tge TriggerEvent) {
			tge.WatchEvent(evtC, wg)
		}(tg)

		// Can't write like this, the next loop tg changed, but go .. is not finished
		// go tg.WatchEvent(evtC, wg)
	}

	go func() {
		for evt := range globalEventC {
			for _, eC := range evtChannls {
				eC <- evt
			}
		}
		for _, eC := range evtChannls {
			close(eC)
		}
	}()
	return
}

func readFWConfig(paths ...string) (fwc FWConfig, err error) {
	for _, cfgPath := range paths {
		data, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			continue
		}
		ext := filepath.Ext(cfgPath)
		switch ext {
		case ".yml":
			if er := yaml.Unmarshal(data, &fwc); er != nil {
				return fwc, er
			}
		case ".json":
			if er := json.Unmarshal(data, &fwc); er != nil {
				return fwc, er
			}
		default:
			err = fmt.Errorf("Unknown format config file: %s", cfgPath)
			return fwc, err
		}
		return fixFWConfig(fwc)
	}
	//fwc, err = fixFWConfig(fwc)
	// data, _ = json.MarshalIndent(fwc, "", "    ")
	// fmt.Println(string(data))
	return fwc, errors.New("Config file not exists")
}

func transformEvent(fsw *fsnotify.Watcher, evtC chan FSEvent) {
	go func() {
		for err := range fsw.Errors {
			log.Errorf("Watch error: %v", err)
		}
	}()
	for evt := range fsw.Events {
		if evt.Op == fsnotify.Create && IsDirectory(evt.Name) {
			log.Info("Add watcher", evt.Name)
			fsw.Add(evt.Name)
			continue
		}
		if evt.Op == fsnotify.Remove {
			if err := fsw.Remove(evt.Name); err == nil {
				log.Info("Remove watcher", evt.Name)
			}
			continue
		}
		if !IsChanged(evt.Name) {
			continue
		}
		//log.Printf("Changed: %s", evt.Name)
		evtC <- FSEvent{ // may panic here
			Name: evt.Name,
		}
	}
}

func initFWConfig() {
	fwc := genFWConfig()
	format := readString("Save format .fsw.(json|yml)", "yml")
	var data []byte
	var cfg string
	if strings.ToLower(format) == "json" {
		data, _ = json.MarshalIndent(fwc, "", "  ")
		cfg = FWCONFIG_JSON
		ioutil.WriteFile(FWCONFIG_JSON, data, 0644)
	} else {
		cfg = FWCONFIG_YAML
		data, _ = yaml.Marshal(fwc)
		ioutil.WriteFile(FWCONFIG_YAML, data, 0644)
	}
	fmt.Printf("Saved to %s\n", strconv.Quote(cfg))
}

func main() {
	version := flag.Bool("version", false, "Show version")
	configfile := flag.String("config", FWCONFIG_YAML, "Specify config file")
	flag.Parse()

	if *version {
		fmt.Println(VERSION)
		return
	}

	subCmd := flag.Arg(0)
	var fwc FWConfig
	var err error
	if subCmd == "" {
		fwc, err = readFWConfig(*configfile, FWCONFIG_JSON)
		if err == nil {
			subCmd = "start"
		} else {
			subCmd = "init"
		}
	}

	switch subCmd {
	case "init":
		initFWConfig()
	case "start":
		visits := make(map[string]bool)
		fsw, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}

		err = WatchPathAndChildren(fsw, fwc.WatchPaths, fwc.WatchDepth, visits)
		if err != nil {
			log.Println(err)
		}

		evtC, wg, err := drainEvent(fwc)
		if err != nil {
			log.Fatal(err)
		}

		sigOS := make(chan os.Signal, 1)
		signal.Notify(sigOS, syscall.SIGINT)
		signal.Notify(sigOS, syscall.SIGTERM)

		go func() {
			sig := <-sigOS
			CPrintf(CPURPLE, "Catch signal %v!", sig)
			close(evtC)
		}()
		go transformEvent(fsw, evtC)
		wg.Wait()
		CPrintf(CPURPLE, "Kill all running ... Done")
	}
}
