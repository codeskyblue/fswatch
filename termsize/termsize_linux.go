// +build linux

package termsize

import (
	"fmt"
	"github.com/aybabtme/color"
)

const (
	TIOCGWINSZ = 0x5413
)

var brush = color.NewBrush("", color.GreenPaint)

func Println(s ...interface{}) {
	p := fmt.Sprint(s...)
	fmt.Println(brush(p))
}
