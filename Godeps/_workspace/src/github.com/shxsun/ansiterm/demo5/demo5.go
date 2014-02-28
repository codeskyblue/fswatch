// demo5.go
// contributed by blamarche

package main

import (
	"ansiterm"
	"fmt"
	"time"
)

func main() {

	//test color/additional functions here
	ansiterm.ResetTerm(ansiterm.NORMAL)
	ansiterm.ClearPage()

	ansiterm.MoveToXY(10, 10)
	sleepOne()

	ansiterm.HideCursor()
	sleepOne()

	ansiterm.SetFGColor(5)
	fmt.Printf("TEST MOVE HIDDEN CURSOR / OUT IN PURPLE")
	sleepOne()

	ansiterm.MoveToXY(10, 12)
	ansiterm.SetColorBright()
	fmt.Printf("BRIGHT COLORS")
	sleepOne()

	ansiterm.MoveToXY(10, 14)
	ansiterm.SetColorNormal()
	fmt.Printf("NORMAL COLORS AGAIN")
	sleepOne()

	ansiterm.ShowCursor()
	sleepOne()

	ansiterm.ResetTerm(ansiterm.NORMAL)
}

func sleepOne() {
	time.Sleep(1000 * time.Millisecond)
}
