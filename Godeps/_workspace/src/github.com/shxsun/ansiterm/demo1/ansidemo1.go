// demo.go

package main

import (
	"ansiterm"
	"flag"
	"fmt"
	"math/rand"
	"time"
)

func pause(n time.Duration) {
	time.Sleep(n * time.Second)
}

func Headline(s string) {
	ansiterm.SavePosn()
	ansiterm.MoveToRC(1, 1)
	ansiterm.ClearLine()
	fmt.Printf("%s", s)
	ansiterm.RestorePosn()
}

func StatusUpdate(s string) {
	const (
		row      = 12
		col      = 4
		fieldlen = 10
	)
	ansiterm.SavePosn()
	ansiterm.MoveToRC(row, col)
	ansiterm.Erase(fieldlen)
	fmt.Printf("%s", s)
	ansiterm.RestorePosn()
}

func test_1() {
	const pauseSec = 2
	fmt.Printf("Test_001\n")
	ansiterm.ResetTerm(ansiterm.NORMAL)
	defer ansiterm.ClearPage()

	ansiterm.ClearPage()
	Headline("this headline should be on row 1")
	StatusUpdate("one")
	pause(pauseSec)

	ansiterm.MoveToRC(3, 1)
	Headline("erase page and print Test on third line")
	StatusUpdate("two")
	fmt.Printf("Test\r")
	pause(pauseSec)

	Headline("erase first 3 chars on third line")
	StatusUpdate("three")
	ansiterm.Erase(3)
	pause(pauseSec)

	Headline("print Best on third line")
	StatusUpdate("four")
	fmt.Printf("Best\r")
	pause(pauseSec)

	Headline("erase entire third line")
	StatusUpdate("five")
	ansiterm.ClearLine()
	pause(pauseSec)

	Headline("print Rest on third line")
	StatusUpdate("six")
	fmt.Printf("Rest\r")
	pause(pauseSec)

	Headline("move to 10,10 and print a msg")
	StatusUpdate("seven")
	ansiterm.MoveToRC(10, 10)
	fmt.Printf("x at 10,10")
	pause(pauseSec)
}

func test_2() {
	const pauseSec time.Duration = 5
	ansiterm.ClearPage()
	ansiterm.MoveToRC(9, 20)
	fmt.Printf("In a few seconds program will print 1000 X chars")
	ansiterm.MoveToRC(10, 20)
	fmt.Printf("This is slowed by forcing a 10 ms sleep per loop")
	pause(pauseSec)
	ansiterm.ClearPage()
	for i := 0; i < 1000; i++ {
		row := int(rand.Int31n(25))
		col := int(rand.Int31n(80))
		ansiterm.MoveToRC(row, col)
		time.Sleep(time.Duration(0.01 * float64(time.Second)))
		fmt.Printf("X")
	}
	pause(pauseSec)
	ansiterm.ClearPage()
	ansiterm.MoveToRC(9, 20)
	fmt.Printf("In a few seconds program will print 1000 X chars")
	ansiterm.MoveToRC(10, 20)
	fmt.Printf("using the same program at FULL speed - don't blink...")
	pause(pauseSec)
	ansiterm.ClearPage()
	for i := 0; i < 1000; i++ {
		row := int(rand.Int31n(25))
		col := int(rand.Int31n(80))
		ansiterm.MoveToRC(row, col)
		fmt.Printf("X")
	}
	pause(pauseSec)
	ansiterm.ClearPage()
}

func main() {
	flag.Parse()
	if flag.NArg() > 0 {
		for i := 0; i < flag.NArg(); i++ {
			fmt.Printf("args(%v)\n", flag.Arg(i))
		}
	}
	test_1()
	test_2()
}
