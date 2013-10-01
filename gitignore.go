package main

import (
	"io/ioutil"
	"regexp"
	"strings"
)

var (
	ignorePattens = make([]*regexp.Regexp, 0, 20)
	ignoreLoaded  = false
)

func loadIgnore() {
	data, err := ioutil.ReadFile(".gitignore")
	if err != nil {
		K.Error(".gitignore file open failed")
		return
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
			ignorePattens = append(ignorePattens, r)
		}
	}
}

func isIgnore(s string) bool {
	if !ignoreLoaded {
		loadIgnore()
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
