// demo3.md

demo3.go
========

Comments can be sent to <hotei1352@gmail.com> . 

---

Under The Hood
--------------

The sensor functions generate fictious readings.  In a production setting they 
would gather real data.  All except for the clock can happen asynchronously which
is provided by the random sleep(). 

The temperature reading is smoothed with a
weighted average since temp can vary quite a bit depending on drive activity.

The seekerr sensor simulates a drive that's getting pushed thru a wide variety
of temperatures and has a lot of seek errors, usually transient in
nature and corrected automatically by the firmware.  You can use smartmontools
to get realtime data.

The clock sensor should be self-explanatory.  If it seems distracting to 
have it tick every second, just change sleepytime to a larger value.

All the sensors run in their own goroutine and are only killed when the program
exits.  There is no way to stop or restart the sensors once going, but you can
change the visibility of their output by changing Field.IsVisible to false.

A production program could be written to use the visibility attribute - perhaps
toggled by a key that calls up a different "View" of the data.

Another thing a production program would likely do is combine the two
types so that it wouldn't be possible to get sensors and fields out of sync.

	type SensorField struct {
		Ch chan int
		DataField Field
	}
	
	temp := newSensorField()
	temp.Ch = make(chan int)
	temp.DataField = {"Temp: ",10,10}
	
	case t := <- temp.Ch: temp.DataField.Show()
	
Initially I thought the array method might be easier to understand, but now that
I've written a few lines of the alternative - I'm not so sure.  See demo4 for code.


// EOF demo3.md  (this is a markdown document and tested OK with blackfriday)
