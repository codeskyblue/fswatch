// utils.go
package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/shxsun/fswatch/termsize"
)

// center string in center(this is a good string)
func StringCenter(s string, count int, padding ...string) string {
	c := "-"
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

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func Md5sum(data []byte) string {
	h := md5.New()
	_, err := h.Write(data)
	if err != nil {
		log.Println("Md5sum error: %s", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// still have space to accelerate
func Md5sumFile(filename string) (sum string, err error) {
	fd, err := os.Open(filename)
	if err != nil {
		return
	}
	h := md5.New()
	_, err = io.Copy(h, fd)
	if err != nil {
		return
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
