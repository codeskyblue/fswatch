// demo4.go
// (c) 2012 David Rook

package main

import (
	"ansiterm"
	"fmt"
	"math/rand"
	"os"
	"time"
)

var license = "(c) 2012 David Rook - released under Simplified FreeBSD license"

var runfor = 1 // minutes to run demo

type Field struct {
	Tag       string
	Data      string
	Row, Col  int
	IsVisible bool
}

type SensorField struct {
	Ch        chan int
	DataField Field
}

var (
	clock    *SensorField
	seekerrs *SensorField
	temp     *SensorField
)

func init() {

}

func tempSensor(t chan int) {
	tempreading := 32
	for {
		sleepytime := rand.Int31n(10)
		time.Sleep(time.Duration(sleepytime) * time.Second)
		newtemp := int(rand.Int31n(100) + 32)
		tempreading = (tempreading*95 + newtemp*5) / 100
		t <- tempreading
	}
}

func seekerrSensor(s chan int) {
	totalerrs := 0
	for {
		sleepytime := 1
		time.Sleep(time.Duration(sleepytime) * time.Second)
		if rand.Int31n(2) == 1 {
			totalerrs++
		}
		s <- totalerrs
	}
}

func clockSensor(c chan int) {
	ticks := 0
	for {
		sleepytime := 1
		time.Sleep(time.Duration(sleepytime) * time.Second)
		ticks++
		c <- ticks
	}
}

func startSensors() {
	go tempSensor(temp.Ch)
	go clockSensor(clock.Ch)
	go seekerrSensor(seekerrs.Ch)
	fmt.Printf("startSensors() fini\n")
}

func initFields() {
	temp = new(SensorField)
	temp.Ch = make(chan int)
	temp.DataField = Field{"Drive Temp(F): ", "", 10, 10, true}

	seekerrs = new(SensorField)
	seekerrs.Ch = make(chan int)
	seekerrs.DataField = Field{"Count seek errors: ", "", 11, 10, true}

	clock = new(SensorField)
	clock.Ch = make(chan int)
	clock.DataField = Field{"TickTock: ", "", 12, 10, true}

	fmt.Printf("initFields() fini\n")
}

// leaves cursor at end of last field updated
func (f *SensorField) Show(s string) {
	ansiterm.MoveToRC(f.DataField.Row, f.DataField.Col)
	ansiterm.Erase(len(f.DataField.Tag) + len(f.DataField.Data))
	f.DataField.Data = s
	fmt.Printf("%s%s", f.DataField.Tag, f.DataField.Data)
}

func main() {
	ansiterm.ResetTerm(0)
	ansiterm.ClearPage()
	initFields()
	startSensors()
	runfor = 1
	done := time.After(time.Duration(runfor) * time.Minute)
	fmt.Printf("This demo will stop after %d minutes\n", runfor)

L1:
	for {
		select {
		case t := <-temp.Ch:
			temp.Show(fmt.Sprintf("%d", t))
		case t := <-clock.Ch:
			clock.Show(fmt.Sprintf("%d", t))
		case t := <-seekerrs.Ch:
			seekerrs.Show(fmt.Sprintf("%d", t))

		// note: label is required else it 'breaks' the select, not the for
		case _ = <-done:
			break L1
		}
	}
	if false {
		os.Exit(0)
	}
	ansiterm.ClearPage()
}
