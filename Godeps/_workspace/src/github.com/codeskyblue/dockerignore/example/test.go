package main

import (
	"bytes"
	"io/ioutil"
	"log"

	ignore "github.com/codeskyblue/dockerignore"
)

func main() {
	// patterns, err := ignore.ReadIgnoreFile(".gitignore")
	rd := ioutil.NopCloser(bytes.NewBufferString("*.exe"))
	patterns, err := ignore.ReadIgnore(rd)
	if err != nil {
		log.Fatal(err)
	}
	isSkip, err := ignore.Matches("hello.exe", patterns)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Should skipped true, got %v", isSkip)
}
