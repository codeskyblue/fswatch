package main

import (
	"github.com/howeyc/fsnotify"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
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
			line = "^" + line + "$"
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

func isLangSource(path string) bool {
	basename := filepath.Base(path)
	if !strings.HasPrefix(basename, ".") {
		ext := filepath.Ext(basename)
		for _, okext := range LangExts {
			if ext == okext {
				return true
			}
		}
	}
	return false
}

func isIgnore(s string) bool {
	if !ignoreLoaded {
		ignorePattens = loadGitignore(".gitignore")
		ignoreLoaded = true
	}
	ok := false
	for _, patten := range ignorePattens {
		if patten.MatchString(s) {
			K.Debugf("patten %s match %s",
				strconv.Quote(patten.String()),
				strconv.Quote(s))
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
			// only lang src can pass
			if !isLangSource(ev.Name) {
				K.Debugf("not source file: %s", ev.Name)
				continue
			}

			// delete or rename has no modify time
			if ev.IsDelete() || ev.IsRename() {
				filterd <- ev
				continue
			}

			fi, err := getFileInfo(ev.Name)
			if err != nil {
				//K.Warnf("get file mod time failed: %s", err)
				continue
			}
			if fi.IsDir() { // ignore directory changes
				K.Debugf("Dir ignore: %s", ev.Name)
				continue
			}

			mt := fi.ModTime().Unix()
			if mt == modifyTime[ev.Name] {
				K.Debugf("SKIP: %s", ev.Name)
				continue
			}

			filterd <- ev
			modifyTime[ev.Name] = mt
		}
	}()
	return filterd
}
