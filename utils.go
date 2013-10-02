// utils.go
package main

import (
	"os"
)

// get file modify time(unix time stamp)
func getFileModTime(path string) (timestamp int64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return
	}
	timestamp = fi.ModTime().Unix()
	return
}
