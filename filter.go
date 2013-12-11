package main

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/howeyc/fsnotify"
)

var (
	ignorePattens []*regexp.Regexp
	ignoreLoaded  = false
)

func loadGitignore(filename string) []*regexp.Regexp {
	ignores := make([]*regexp.Regexp, 0, 20)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		logs.Infof("file '%s' open failed", filename)
		return nil
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if !strings.HasPrefix(line, "#") {
			line = strings.Replace(line, ".", "[.]", -1)
			line = strings.Replace(line, "*", ".*", -1)
			line = "^" + line + "$"
			r, err := regexp.Compile(line)
			if err != nil {
				logs.Error(err)
				continue
			}
			logs.Debug("init: ", line)
			ignores = append(ignores, r)
		}
	}
	return ignores
}

type Action int

const (
	FALLTHROUGH = Action(iota)
	REJECT
	ACCEPT
)

type FilterFunc func(*fsnotify.FileEvent) Action

func extentionFilter(ev *fsnotify.FileEvent) Action {
	if len(LangExts) == 0 {
		return FALLTHROUGH
	}
	ext := filepath.Ext(filepath.Base(ev.Name))
	for _, okExt := range LangExts {
		if ext == "."+okExt {
			return ACCEPT
		}
	}
	return REJECT
}

// ignore hidden file
func hiddenFilter(ev *fsnotify.FileEvent) Action {
	basename := filepath.Base(ev.Name)
	if strings.HasPrefix(basename, ".") {
		logs.Warn("hidden file", ev.Name)
		return REJECT
	}
	return FALLTHROUGH
}

func gitignoreFilter(ev *fsnotify.FileEvent) Action {
	s := ev.Name
	if !ignoreLoaded {
		ignorePattens = loadGitignore(".gitignore")
		ignoreLoaded = true
	}
	for _, patten := range ignorePattens {
		if patten.MatchString(s) {
			logs.Debugf("patten %s match %s",
				strconv.Quote(patten.String()),
				strconv.Quote(s))
			return REJECT
		}
	}
	return FALLTHROUGH
}

func filter(watch chan *fsnotify.FileEvent, funcs ...FilterFunc) chan *fsnotify.FileEvent {
	modifyTime := make(map[string]int64)
	filterd := make(chan *fsnotify.FileEvent)
	go func() {
		for {
			ev := <-watch
			break2 := false
			for i, filterFunc := range funcs {
				// FIXME: how to get func name
				//name := reflect.ValueOf(filterFunc).Type().Name()
				//logs.Warn("func name", name)
				n := filterFunc(ev)
				logs.Debug("filter", i, "action:", n)
				if n == FALLTHROUGH {
					continue
				} else if n == REJECT {
					break2 = true
					break
				} else if n == ACCEPT {
					filterd <- ev
					break
				}
			}

			if break2 {
				continue
			}

			// delete or rename has no modify time
			if ev.IsDelete() || ev.IsRename() {
				filterd <- ev
				continue
			}

			//if isIgnore(filepath.Base(ev.Name)) {
			//	logs.Debugf("IGNORE: %s", ev.Name)
			//	continue
			//}

			fi, err := getFileInfo(ev.Name)
			if err != nil {
				//logs.Warnf("get file mod time failed: %s", err)
				continue
			}
			if fi.IsDir() { // ignore directory changes
				logs.Debugf("Dir ignore: %s", ev.Name)
				continue
			}

			mt := fi.ModTime().Unix()
			if mt == modifyTime[ev.Name] {
				logs.Debugf("SKIP: %s", ev.Name)
				continue
			}

			filterd <- ev
			modifyTime[ev.Name] = mt
		}
	}()
	return filterd
}
