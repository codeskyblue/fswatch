package fswatch

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-fsnotify/fsnotify"
)

// 0: indicate current path depth
func WalkDir(path string, depth int) (dirs []string, err error) {
	dirs = make([]string, 0)
	baseNumSeps := strings.Count(path, string(os.PathSeparator))
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			base := info.Name()
			if base != "." && strings.HasPrefix(base, ".") { // ignore hidden dir
				return filepath.SkipDir
			}

			pathDepth := strings.Count(path, string(os.PathSeparator)) - baseNumSeps
			//log.Println(path, pathDepth)
			if depth > 0 && pathDepth > depth {
				return filepath.SkipDir
			}
			dirs = append(dirs, path)
		}
		return nil
	})
	return
}

func Watch(files []string, ignores []string) (evC chan Event, err error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}

	evC = make(chan Event)
	for _, file := range files {
		fsw.Add(file)
	}
	go func() {
		for {
			select {
			case event := <-fsw.Events:
				evC <- Event{
					Name: event.Name,
					Type: ET_FILESYSTEM,
					Op:   Op(event.Op),
				}
			case fserr := <-fsw.Errors:
				log.Println(fserr)
			}
		}
	}()
	return evC, nil
}
