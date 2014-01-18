// +build !windows

package termsize

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/aybabtme/color"
)

type winsize struct {
	ws_row, ws_col       uint16
	ws_xpixel, ws_ypixel uint16
}

func getWS() *winsize {
	ws := &winsize{}
	if TIOCGWINSZ != 0 {
		syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(0),
			uintptr(TIOCGWINSZ),
			uintptr(unsafe.Pointer(ws)))
		return ws
	}
	// other
	ws.ws_col = 80
	ws.ws_row = 40
	return ws
}

func Width() int {
	return int(getWS().ws_col)
}

func Height() int {
	return int(getWS().ws_row)
}

func GetTerminalColumns() int {
	ws := winsize{}

	if TIOCGWINSZ != 0 {
		syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(0),
			uintptr(TIOCGWINSZ),
			uintptr(unsafe.Pointer(&ws)))

		return int(ws.ws_col)
	}

	return 80
}

var brush = color.NewBrush("", color.GreenPaint)

func Println(s ...interface{}) {
	p := fmt.Sprint(s...)
	fmt.Println(brush(p))
}
