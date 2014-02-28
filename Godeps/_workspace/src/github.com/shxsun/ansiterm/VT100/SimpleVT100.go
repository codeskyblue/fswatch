// SimpleVT100.go

/*
 * (c) 2010 David Rook
 */

package main

import (
	"fmt"
	"os"
)

const (
	ESC   = 033
	BEL   = 0x07
	COLOR = false
	NROW  = 25
	NCOL  = 80
)

// 0 black
// 1 red
// 2 green
// 3 orangish bkg
// 4 navy blue
// 5 purple
// 6 cyan
// 7 grey bkg
// 8 cyan
// 9 white

const (
	BLACK = iota
	RED
	GREEN
	ORANGE
	NAVYBLUE
	PURPLE
	CYAN
	GREY
	WHITE = 9
	GRAY  = GREY
)

var (
	cfcolor int  = BLACK
	cbcolor int  = WHITE
	state   bool = true
)

// set foreground color
func ansifcolor(color int) {
	if color == cfcolor {
		return
	}
	color &= 0x7
	ttputc(ESC)
	ttputc('[')
	ansiparm(color + 30)
	ttputc('m')
	cfcolor = color
}

// set background color
func ansibcolor(color int) {
	if color == cbcolor {
		return
	}
	color &= 0x7
	ttputc(ESC)
	ttputc('[')
	ansiparm(color + 40)
	ttputc('m')
	cbcolor = color
}

func ansirev(state bool) {
	ttputc(ESC)
	ttputc('[')
	if state {
		ttputc('7')
	} else {
		ttputc('0')
	}
	ttputc('m')
	ftmp := cfcolor
	btmp := cbcolor
	ansifcolor(ftmp)
	ansibcolor(btmp)
	state = !state
}

// output one character
func ttputc(c byte) {
	b := make([]byte, 1) // TODO move this to global var?
	b[0] = c
	n, err := os.Stdout.Write(b)
	if err != nil {
	} // TODO
	n = n // TODO
}

// erase to end of line
func ansieeol() {
	ansifcolor(cfcolor)
	ansibcolor(cbcolor)
	ttputc(ESC)
	ttputc('[')
	ttputc('K')
}

// erase to end of page
func ansieeop() {
	ansifcolor(cfcolor)
	ansibcolor(cbcolor)
	ttputc(ESC)
	ttputc('[')
	ttputc('J')
}

func ansiparm(n int) {
	var q, r int

	q = n / 10
	if q != 0 {
		r = q / 10
		if r != 0 {
			ttputc(byte((r % 10) + '0'))
		}
		ttputc(byte((q % 10) + '0'))
	}
	ttputc(byte((n % 10) + '0'))
}

func ansimoverc(row, col int) {
	ttputc(ESC)
	ttputc('[')
	ansiparm(row + 1)
	ttputc(';')
	ansiparm(col + 1)
	ttputc('H')
}

func ansimovexy(x, y int) {
	ansimoverc(y, x)
}

func ansibeep() {
	ttputc(BEL)
	//	ttflush();
}

func test() {
	ansibcolor(WHITE)
	ansifcolor(GREY)
	ansimovexy(0, 0)
	ansieeop()

	col := 7
	for row := 0; row < 16; row++ {
		ansimoverc(row, col)
		ansifcolor(row)
		ttputc('X')
		col++
	}
	for row := 16; row >= 0; row-- {
		ansimoverc(row, col)
		col++
		ttputc('X')
	}
	ansimovexy(0, 3)
	ansibcolor(GREEN)
	ansieeol()
	ansimovexy(0, 20)
	ansibcolor(WHITE)
	ansifcolor(BLACK)
}

func main() {
	fmt.Printf("SimpleVT100.go\n")
	test()
	fmt.Printf("White is %d\n", WHITE)

}
