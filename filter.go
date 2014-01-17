package main

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/howeyc/fsnotify"
)

var (
	ignorePattens []*regexp.Regexp
	ignoreLoaded  = false
)

type Action int

const (
	ACTION_ACCEPT = Action(iota)
	ACTION_REJECT
	ACTION_CONTINUE
)

// load .gitignore
func loadGitignore(filename string) []*regexp.Regexp {
	ignores := make([]*regexp.Regexp, 0, 20)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		logs.Debugf("file '%s' open failed", filename)
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

type FilterFunc func(*fsnotify.FileEvent) Action

func extentionFilter(ev *fsnotify.FileEvent) Action {
	if len(LangExts) == 0 {
		return ACTION_CONTINUE
	}
	ext := filepath.Ext(filepath.Base(ev.Name))
	for _, okExt := range LangExts {
		if ext == "."+okExt {
			return ACTION_CONTINUE
		}
	}
	return ACTION_REJECT
}

// ignore hidden file
func hiddenFilter(ev *fsnotify.FileEvent) Action {
	basename := filepath.Base(ev.Name)
	if strings.HasPrefix(basename, ".") {
		return ACTION_REJECT
	}
	return ACTION_CONTINUE
}

// if delete or rename, just accept
func deleteRenameFilter(ev *fsnotify.FileEvent) Action {
	if ev.IsDelete() || ev.IsRename() {
		return ACTION_REJECT
	}
	return ACTION_CONTINUE
}

var md5sumMap = make(map[string]string)
var mapLock = &sync.Mutex{}

// call after deleteRename
func md5CheckFileter(ev *fsnotify.FileEvent) Action {
	mapLock.Lock()
	defer mapLock.Unlock()
	name := ev.Name
	sum, err := Md5sumFile(name)
	if err != nil {
		logs.Error(err)
		return ACTION_CONTINUE
	}
	logs.Debugf("md5file: %s, sum: %s", name, sum)
	oldsum, exists := md5sumMap[name]
	if sum != oldsum {
		md5sumMap[name] = sum
		_ = exists
		//if !exists {
		//	return ACTION_REJECT
		//}
		return ACTION_ACCEPT
	}
	return ACTION_REJECT
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
			return ACTION_REJECT
		}
	}
	return ACTION_CONTINUE
}

func filter(watch chan *fsnotify.FileEvent, funcs ...FilterFunc) chan *fsnotify.FileEvent {
	modifyTime := make(map[string]int64)
	filterd := make(chan *fsnotify.FileEvent)
	go func() {
		for {
		AGAIN:
			ev := <-watch
			for _, filterFunc := range funcs {
				n := filterFunc(ev)
				logs.Debugf("filter func: %s action:%d",
					GetFunctionName(filterFunc), n)
				//logs.Debugf("event: %s", ev)
				if n == ACTION_CONTINUE {
					continue
				} else if n == ACTION_REJECT {
					goto AGAIN
					//break2 = true
					break
				} else if n == ACTION_ACCEPT {
					filterd <- ev
					break
				}
			}

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
