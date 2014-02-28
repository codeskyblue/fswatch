// utils.go
package klog

import (
	"os"
	"runtime"
)

func isTermOutput() (result bool) {
	switch runtime.GOOS {
	case "linux", "darwin":
		fi, _ := os.Stdout.Stat()
		return fi.Mode()&os.ModeCharDevice == os.ModeCharDevice
	case "windows":
		return os.Stdout.Fd() < 0xff
	}
	return false
}

//method, exists := l.color.getMethod(colorName)
//if exists {
//	outstr = method.Func.Call([]reflect.Value{
//		reflect.ValueOf(l.color),
//		reflect.ValueOf(outstr)},
//	)[0].String()
//}
