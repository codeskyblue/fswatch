// utils.go
package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
)

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
	h.Write(data)
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
