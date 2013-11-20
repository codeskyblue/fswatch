// +build linux

package termsize

import (
	"fmt"
	"github.com/aybabtme/color"
)

const (
	TIOCGWINSZ = 0x5413
)

func Println(s ...interface{}) {
	var brush = color.NewBrush("", color.DarkGreenPaint)
	p = fmt.Sprintln(s...)
	fmt.Println(brush(p))
}
