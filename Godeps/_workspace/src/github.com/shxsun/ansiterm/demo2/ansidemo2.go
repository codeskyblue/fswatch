// ansidemo2.go

package main

import (
	"ansiterm"
	"flag"
	"fmt"
	"math/rand"
	"time"
)

type Ipt struct {
	row, col int
}

const (
	snakeLen = 16
	maxRow   = 25
	maxCol   = 80
)

type Snake struct {
	body       []Ipt
	tail       Ipt
	dirx, diry int
}

func flipcoin() bool {
	if rand.Int31n(2) == 1 {
		return true
	}
	return false
}

func (s *Snake) Show() {
	// erase tail
	ansiterm.MoveToRC(s.tail.row, s.tail.col)
	fmt.Printf(" ")
	// print body
	for _, rib := range s.body {
		ansiterm.MoveToRC(rib.row, rib.col)
		fmt.Printf("s")
	}
}

func (s *Snake) Slither() {
	s.tail = s.body[0]
	newrib := s.body[len(s.body)-1]
	if flipcoin() {
		newrib.row += s.diry
	}
	if flipcoin() {
		newrib.col += s.dirx
	}
	if newrib.col >= maxCol {
		newrib.col = maxCol
		s.dirx = -s.dirx
	}
	if newrib.row >= maxRow {
		newrib.row = maxRow
		s.diry = -s.diry
	}
	if newrib.col <= 1 {
		newrib.col = 1
		s.dirx = -s.dirx
	}
	if newrib.row <= 1 {
		newrib.row = 1
		s.diry = -s.diry
	}
	// chop off tail, add newrib at head
	s.body = append(s.body[1:], newrib)
}

func pause() {
	time.Sleep(time.Duration(0.01 * float64(time.Second)))
}

func test_1() {
	s := new(Snake)
	s.dirx = 1
	s.diry = 1
	row := 10
	col := 10
	s.tail = Ipt{row, col}
	for i := 0; i < snakeLen; i++ {
		s.body = append(s.body, Ipt{row, col})
		row++
	}

	for {
		s.Slither()
		s.Show()
		pause()
	}
}

func main() {
	flag.Parse()
	if flag.NArg() > 0 {
		for i := 0; i < flag.NArg(); i++ {
			fmt.Printf("args(%v)\n", flag.Arg(i))
		}
	}
	_ = rand.Int31n(10)
	ansiterm.ClearPage()
	test_1()
}
