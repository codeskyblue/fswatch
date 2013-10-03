package main

import (
	"github.com/howeyc/fsnotify"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	ignorePattens []*regexp.Regexp
	ignoreLoaded  = false
)

func loadGitignore(filename string) []*regexp.Regexp {
	ignores := make([]*regexp.Regexp, 0, 20)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		K.Errorf("file '%s' open failed", filename)
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
			r, err := regexp.Compile(line)
			if err != nil {
				K.Error(err)
				continue
			}
			K.Debug("init: ", line)
			ignores = append(ignores, r)
		}
	}
	return ignores
}

func isIgnore(s string) bool {
	if !ignoreLoaded {
		ignorePattens = loadGitignore(".gitignore")
		ignoreLoaded = true
	}
	ok := false
	for _, patten := range ignorePattens {
		if patten.MatchString(s) {
			ok = true
			break
		}
	}
	return ok
}

func filter(watch chan *fsnotify.FileEvent) chan *fsnotify.FileEvent {
	modifyTime := make(map[string]int64)
	filterd := make(chan *fsnotify.FileEvent)
	go func() {
		for {
			ev := <-watch
			if isIgnore(filepath.Base(ev.Name)) {
				K.Debugf("IGNORE: %s", ev.Name)
				continue
			}
			// delete or rename has no modify time
			if ev.IsDelete() || ev.IsRename() {
				filterd <- ev
				continue
			}
			mt, err := getFileModTime(ev.Name)
			if err != nil {
				//K.Warnf("get file mod time failed: %s", err)
				continue
			}
			t := modifyTime[ev.Name]
			if mt == t {
				K.Debugf("SKIP: %s", ev.Name)
				continue
			}

			filterd <- ev
			modifyTime[ev.Name] = mt
		}
	}()
	return filterd
}
