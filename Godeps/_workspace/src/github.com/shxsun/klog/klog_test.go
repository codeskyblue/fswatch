package klog

import (
	"bytes"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
)

var K *Logger = NewLogger(nil, "i am prefix ")
var out = bytes.NewBuffer(nil)
var kt *Logger = NewLogger(out, "")

func TestNewFileLogger(t *testing.T) {
	filename := "testdata/tmpfile"
	log, err := NewFileLogger(filename)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		os.Remove(filename)
	}()
	log.Info("hello")
	fd, err := os.Open(filename)
	if err != nil {
		t.Error(err)
	}
	stat, _ := fd.Stat()
	if stat.Size() <= 0 {
		t.Error("expect write into something, but klog write nothing")
	}
}

func TestDebugf(t *testing.T) {
	out := bytes.NewBuffer(nil)
	k := NewLogger(out, "")
	k.SetLevel(LDebug)
	k.Debugf("hello %s", "klog")
	outStr := string(out.Bytes())
	if !strings.Contains(outStr, "hello klog") {
		t.Errorf("expect suffix with %s, but receive %s",
			strconv.Quote("hello klog"),
			strconv.Quote(outStr))
	}
}

func TestAll(t *testing.T) {
	K.SetLevel(LDebug)
	K.Debug("this is debug")
	K.Info("this is info")
	K.Warn("this is warn")
	K.Error("this is error")
	//K.Fatal("msg:fatal")
}

func TestNoColor(t *testing.T) {
	flags := K.Flags()
	K.SetFlags(flags & ^Fcolor)
	K.Info("this info msg has no color")
}

func TestSetLevel(t *testing.T) {
	out.Reset()
	kt.SetLevel(LInfo)
	kt.Debug("dddd")
	if len(out.Bytes()) != 0 {
		t.Error("expect empty, but got sth")
	}
	kt.Info("iiii")
	if len(out.Bytes()) == 0 {
		t.Error("expect output sth, but nothing got")
	}
	out.Reset()
	kt.Error("eeee")
	if len(out.Bytes()) == 0 {
		t.Error("expect output sth, but nothing got")
	}
}

func BenchmarkTest(b *testing.B) {
	b.StopTimer()
	k := NewLogger(ioutil.Discard, "")
	b.StartTimer()
	for i := 0; i < 100; i++ {
		k.Debug("ddddddddddddddddddd", "wwwwwwwwwwwwwwwwwwww")
		k.Debug("ddddddddddddddddddd", "wwwwwwwwwwwwwwwwwwww")
		k.Debug("ddddddddddddddddddd", "wwwwwwwwwwwwwwwwwwww")
		k.Debug("ddddddddddddddddddd", "wwwwwwwwwwwwwwwwwwww")
		k.Debug("ddddddddddddddddddd", "wwwwwwwwwwwwwwwwwwww")
		k.Info("ddddddddddddddddddd", "wwwwwwwwwwwwwwwwwwww")
		k.Info("ddddddddddddddddddd", "wwwwwwwwwwwwwwwwwwww")
		k.Info("ddddddddddddddddddd", "wwwwwwwwwwwwwwwwwwww")
		k.Warn("ddddddddddddddddddd", "wwwwwwwwwwwwwwwwwwww")
		k.Error("ddddddddddddddddddd", "wwwwwwwwwwwwwwwwwwww")
	}
}

func TestDefaultLog(t *testing.T) {
	DevLog.Debug("dev debug will show file:line")
	DevLog.Warn("dev warn")
	//DevLog.Fatal("fatal")
	StdLog.Debug("std debug will not showen")
	StdLog.Warn("warn message will show")
}
