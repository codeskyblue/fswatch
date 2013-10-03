// utils.go
package main

import (
	"os"
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
