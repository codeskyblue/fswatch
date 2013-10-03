package main

import (
	"io/ioutil"
	"os"
	"testing"
)

var sample = `
# comment
#
*.cgo
filter.exe

#
*.swx
*.swp
`

func TestFilter(t *testing.T) {
	testdata := "testdata/sample.ignore"
	err := ioutil.WriteFile(testdata, []byte(sample), 0644)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(testdata)
	ignorePattens = loadGitignore("testdata/sample.ignore")
	if isIgnore("xxx.go") {
		t.Error("xxx.go should not be ignored")
	}
	if !isIgnore("xxx.swp") {
		t.Error("xxx.swp should be ignored")
	}
	if !isIgnore("filter.exe") {
		t.Error("filter.exe should be ignored")
	}
}
