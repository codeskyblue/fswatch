// +build !windows

package termsize

import (
	"fmt"
	"github.com/aybabtme/color"
	"syscall"
	"unsafe"
)

type winsize struct {
	ws_row, ws_col       uint16
	ws_xpixel, ws_ypixel uint16
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
