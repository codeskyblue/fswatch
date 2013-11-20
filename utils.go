// utils.go
package main

import (
	"github.com/shxsun/fswatch/termsize"
	"os"
	"strings"
)

// center string in center
func StringCenter(s string, count int, padding ...string) string {
	c := "="
	if len(padding) != 0 {
		c = padding[0]
	}
	tot := count - len(s)
	if tot <= 0 {
		return s
	}
	left, right := tot/2, (tot+1)/2
	return strings.Repeat(c, left) + s + strings.Repeat(c, right)
}

var TermSize int

func init() {
	TermSize = termsize.GetTerminalColumns()
}

func getFileInfo(path string) (fi os.FileInfo, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	fi, err = f.Stat()
	return
}
