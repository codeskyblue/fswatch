// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9,!solaris

// Package fsnotify provides a platform-independent interface for file system notifications.
package fswatch

import (
	"bytes"
	"fmt"
)

// Event represents a single file system notification.
type Event struct {
	Name string // Relative path to the file or directory.
	Type EventType
	Op   Op // File operation that triggered the event.
}

// Op describes a set of file operations.
type Op uint32

// These are the generalized file operations that can trigger a notification.
const (
	Create Op = 1 << iota
	Write
	Remove
	Rename
	Chmod
)

type EventType uint32

const (
	ET_SIGNAL EventType = 1 << iota
	ET_FILESYSTEM
)

// String returns a string representation of the event in the form
// "file: REMOVE|WRITE|..."
func (e Event) String() string {
	if e.Type == ET_SIGNAL {
		return fmt.Sprintf("Signal event: %v", e.Name)
	}
	// Use a buffer for efficient string concatenation
	var buffer bytes.Buffer

	if e.Op&Create == Create {
		buffer.WriteString("|CREATE")
	}
	if e.Op&Remove == Remove {
		buffer.WriteString("|REMOVE")
	}
	if e.Op&Write == Write {
		buffer.WriteString("|WRITE")
	}
	if e.Op&Rename == Rename {
		buffer.WriteString("|RENAME")
	}
	if e.Op&Chmod == Chmod {
		buffer.WriteString("|CHMOD")
	}

	// If buffer remains empty, return no event names
	if buffer.Len() == 0 {
		return fmt.Sprintf("%q: ", e.Name)
	}

	// Return a list of event names, with leading pipe character stripped
	return fmt.Sprintf("%q: %s", e.Name, buffer.String()[1:])
}
