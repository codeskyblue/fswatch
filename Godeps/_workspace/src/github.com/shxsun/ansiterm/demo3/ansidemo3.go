// demo3.go  (c) 2012-2013 David Rook

// note use "$ reset" to recover cursor if it is hidden on exit

package main

import (
	"ansiterm"
	"fmt"
	"math/rand"
	//"os"
	"time"
)

var license = "(c) 2012 David Rook - released under Simplified FreeBSD license"

var sensors []chan int

var g_Form ansiterm.ScreenForm

var runfor = 1 // minutes to run demo

func init() {
	sensors = make([]chan int, 0)
}

// generates a random temp reading between 32 and 132, then does 95% weighted avg
// long term reading should stabilize around 82 degrees
// readings appear at random with average rate of half the max sleep interval
func tempSensor(t chan int) {
	tempreading := 32
	for {
		sleepytime := rand.Int31n(4)
		time.Sleep(time.Duration(sleepytime) * time.Second)
		newtemp := int(rand.Int31n(100) + 32)
		tempreading = (tempreading*95 + newtemp*5) / 100
		t <- tempreading
	}
}

// generates seek error count that increases randomly at average of 30 per minute
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

// generates a clock tick every second
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
	t := make(chan int)
	sensors = append(sensors, t)
	go tempSensor(t)

	s := make(chan int)
	sensors = append(sensors, s)
	go seekerrSensor(s)

	c := make(chan int)
	sensors = append(sensors, c)
	go clockSensor(c)

	fmt.Printf("startSensors() fini\n")
}

func setupForm() {

	{
		x := new(ansiterm.ScreenField)
		x.SetTag("title")
		x.SetRCW(2, 35, 15)
		x.SetPrompt("ansiterm Demo 3")
		g_Form.AddField(x)
	}
	x := new(ansiterm.ScreenField)
	x.SetTag("temp")
	x.SetRCW(10, 4, 21)
	x.SetPrompt("Drive Temp: ")
	g_Form.AddField(x)

	y := new(ansiterm.ScreenField)
	y.SetTag("seekerr")
	y.SetRCW(11, 6, 15)
	y.SetPrompt("SeekErrs: ")
	g_Form.AddField(y)

	z := new(ansiterm.ScreenField)
	z.SetTag("time")
	z.SetRCW(12, 10, 15)
	z.SetPrompt("Time: ")
	g_Form.AddField(z)

	fmt.Printf("setupForm() fini\n")
}

func progressBar(prompt string, row, col, barWidth int, p chan int, alldone int) {
	var progress int
	z := new(ansiterm.ScreenField)
	z.SetTag("progress")
	z.SetPrompt(prompt)
	z.SetRCW(row, col, barWidth+len(prompt))
	g_Form.AddField(z)
	fmt.Printf("Progress bar started\n")
	bar := "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
	nobar := "...................................................................."
	g_Form.UpdateMsg("progress", nobar)
	for {
		progress = <-p
		var barlen int = (progress * barWidth) / alldone
		s := bar[:barlen] + nobar[:barWidth-barlen]
		g_Form.UpdateMsg("progress", s)
	}
}

func main() {
	ansiterm.ResetTerm(0)
	ansiterm.ClearPage()
	//	ansiterm.MoveToRC(13,13)
	//	ansiterm.QueryPosn()  hangs
	//	os.Exit(0)
	ansiterm.HideCursor()
	defer ansiterm.ShowCursor()
	setupForm()
	startSensors()
	runfor = 1
	prog := make(chan int)
	// prompt, row,col,barWidth,listen chan, alldone
	go progressBar("Progress: ", 20, 20, 50, prog, 60*runfor)
	done := time.After(time.Duration(runfor) * time.Minute) // set up a timeout channel
	fmt.Printf("This demo will stop after %d minute(s)\n", runfor)
	time.Sleep(2 * time.Second)
L1:
	for {
		select {
		case t := <-sensors[0]:
			g_Form.UpdateMsg("temp", fmt.Sprintf("%5d", t))
		case t := <-sensors[1]:
			g_Form.UpdateMsg("seekerr", fmt.Sprintf("%5d", t))
		case t := <-sensors[2]:
			g_Form.UpdateMsg("time", fmt.Sprintf("%5d", t))
			g_Form.UpdateMsg("progress", fmt.Sprintf("%5d", t))
			prog <- t
			g_Form.Draw()
		// note: label is required else it 'breaks' the select, not the for
		case _ = <-done:
			break L1
		}
	}
	if true { // optional wrapup stats
		ansiterm.ClearPage()
		ansiterm.ResetTerm(ansiterm.NORMAL)
		fmt.Printf("<done>\n")
	}
}
