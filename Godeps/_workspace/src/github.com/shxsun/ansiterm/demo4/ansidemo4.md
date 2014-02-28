// ansidemo4.md

demo4.go
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
Based on the math, one would expect the temp to stabilize somewhere around
82 degrees with a range from [32..132].

The seekerr sensor simulates a drive that's getting pushed thru a wide variety
of temperatures and has a lot of seek errors, usually transient in
nature and corrected automatically by the firmware.

The clock sensor should be self-explanatory.  If it seems distracting to 
have it tick every second, just change sleepytime to a larger value.

All the sensors run in their own goroutine and are only killed when the program
exits.  There is no way to stop or restart the sensors once going, but you can
change the visibility of their output by changing Field.IsVisible to false.

A production program could be written to use the visibility attribute - perhaps
toggled by a key that calls up a different "View" of the data.  

// EOF ansidemo4.md  (this is a markdown document and tested OK with blackfriday)
