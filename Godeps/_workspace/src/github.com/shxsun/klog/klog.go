// beelog project klog
package klog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/aybabtme/color"
)

type Level int

const (
	Fshortfile = 1 << iota // show filename:lineno
	Fdate
	Ftime
	Fcolor

	Fdatetime = Fdate | Ftime
	Fdevflag  = Fdatetime | Fshortfile | Fcolor // for develop use
	Fstdflag  = Fdatetime | Fcolor
)

const (
	LDebug Level = iota
	LInfo
	LWarning
	LError
	LFatal
)

var (
	DevLog = NewLogger(os.Stdout, "").SetFlags(Fdevflag).SetLevel(LDebug)
	StdLog = NewLogger(os.Stderr, "").SetFlags(Fstdflag).SetLevel(LInfo)
)

var levels = []string{
	"[DEBUG]",
	"[INFO]",
	"[WARN]",
	"[ERROR]",
	"[FATAL]",
}

var colors = []color.Paint{
	color.CyanPaint,
	color.GreenPaint,
	color.YellowPaint,
	color.RedPaint,
	color.PurplePaint,
}

var mu = &sync.Mutex{}

type Logger struct {
	out         io.Writer
	level       Level
	writer      io.Writer
	flags       int
	prefix      string
	colorEnable bool
}

// Default LogLevel Debug
func NewFileLogger(filename string) (log *Logger, err error) {
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return
	}
	return NewLogger(fd).SetFlags(Fdevflag).SetLevel(LDebug), nil
}

// default level is debug
func NewLogger(out io.Writer, prefix ...string) *Logger {
	if out == nil {
		out = os.Stdout
	}
	return &Logger{
		level:       LInfo,
		writer:      out,
		colorEnable: runtime.GOOS != "windows" && isTermOutput(), // TODO: isTemOutput is not so good
		flags:       Fstdflag,
		prefix:      strings.Join(prefix, " "),
	}
}

// set flags to change klog output style
func (l *Logger) SetFlags(flag int) *Logger {
	l.flags = flag
	return l
}

// get logger flags
func (l *Logger) Flags() int {
	return l.flags
}

// set output level. L[Debug|Warning...]
func (l *Logger) SetLevel(level Level) *Logger {
	l.level = level
	return l
}

// get current level.
func (l *Logger) Level() Level {
	return l.level
}

func (l *Logger) write(level Level, format string, a ...interface{}) (n int, err error) {
	if level < l.level {
		return
	}
	var levelName string = levels[int(level)]
	var sep = " "
	var prefix, outstr = l.prefix, ""

	if l.flags&Fdatetime != 0 {
		now := time.Now()
		layout := ""
		if l.flags&Fdate != 0 {
			layout += "2006/01/02"
		}
		if l.flags&Ftime != 0 {
			layout += " 15:04:05"
		}
		layout = strings.TrimSpace(layout)
		prefix += now.Format(layout)
	}

	if l.flags&Fshortfile != 0 {
		// Retrieve the stack infos
		_, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "<unknown>"
			line = -1
		} else {
			file = file[strings.LastIndex(file, "/")+1:]
		}
		prefix = fmt.Sprintf("%s %s:%d", prefix, file, line)
	}

	outstr += levelName

	if format == "" {
		for _, i := range a {
			outstr += sep + fmt.Sprintf("%v", i)
		}
	} else {
		outstr = outstr + sep + fmt.Sprintf(format, a...)
	}
	if !strings.HasSuffix(outstr, "\n") {
		outstr += "\n"
	}

	if l.colorEnable && l.flags&Fcolor != 0 {
		brush := color.NewBrush("", colors[int(level)])
		outstr = brush(outstr)
	}

	mu.Lock()
	defer mu.Unlock()
	return l.writer.Write([]byte(prefix + sep + outstr))
}

func (l *Logger) Debug(v ...interface{}) {
	l.write(LDebug, "", v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.write(LDebug, format, v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.write(LInfo, "", v...)
}
func (l *Logger) Infof(format string, v ...interface{}) {
	l.write(LInfo, format, v...)
}
func (l *Logger) Warn(v ...interface{}) {
	l.write(LWarning, "", v...)
}
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.write(LWarning, format, v...)
}
func (l *Logger) Error(v ...interface{}) {
	l.write(LError, "", v...)
}
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.write(LError, format, v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.write(LFatal, format, v...)
	os.Exit(1)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.write(LFatal, "", v...)
	os.Exit(1)
}
