
<center>
ansiterm
========
</center>

License details are at the end of this document. 
This document is (c) 2013 David Rook.

Comments can be sent to <hotei1352@gmail.com>

Description
-----------
This [package][1] allows you to control consoles, xterms and other things that respond
to ANSI Escape Code sequences.

---

Installation
------------

```
go get github.com/hotei/ansiterm
```

Under The Hood
--------------

Source for escape sequences is the ANSI Escape Code [Wikipedia article][2].

It's primarily intended to make "status reports" easier to handle on
long running applications that don't warrant a full GUI display.  These
applications happen to be running on servers with a text only console.
For some examples see the demo directory.

For the sake of simplicity I only implemented the functions needed for status reports.

Features
--------
More details at GoDoc.Org http://godoc.org/github.com/hotei/ansiterm

* Save cursor position
* Restore cursor position
* Erase N characters without moving cursor position
* Erase whole line
* Erase whole page
* Move cursor to specific row,col
* Reset terminal

Several other functions were added by blamarche to handle color
see code for details 

You can see the [wiki][1] article mentions many more codes.  I have no plans at
present to expand coverage but it's very easy to do.

As noted in the source I'm not
concerned about this program's efficiency (or rather lack of).
My programs tend to spend 99.9% of the time crunching 
numbers and that last point one percent in I/O.  If your situation is different
then time spent replacing fmt.Printf() could be worthwhile.

Also note the following caveats for MS Windows - again from the [wiki][2] 
<font color=red>"Win32 does not support ANSI sequences at all."</font>

* Won't work for older MSWin consoles without ANSI.SYS ie. DOS 1.0
* Won't work for newer MSWin consoles - Win32 and later 
* Will work on your VT-100 if the hardware still functions :-)
* Works on (some/most?) xterms, including default GNOME and KDE ? consoles
* Works on FreeBSD (and others) using text consoles that respond to ANSI codes.
FreeBSD 9.x default console works ok.


---

Resources
---------
* [Source for package] [1]
* [Wiki Escape Codes] [2]
* [Blamarche Fork] [3]

---

Changes
-------
* 2013-05-08 Doc updated
* 2013-04-10 Code updated from blamarche pull request
* 20?? - started

[1]: http://github.com/hotei/ansiterm "github.com/hotei/ansiterm"
[2]: http://en.wikipedia.org/wiki/ANSI_escape_code "Wiki ANSI_escape_code"
[3]: http://github.com/blamarche/ansiterm "github.com/blamarche/ansiterm"

License
-------
The 'ansiterm' go package and demo programs are distributed under the Simplified BSD License:

> Copyright (c) 2013 David Rook. All rights reserved.
> 
> Redistribution and use in source and binary forms, with or without modification, are
> permitted provided that the following conditions are met:
> 
>    1. Redistributions of source code must retain the above copyright notice, this list of
>       conditions and the following disclaimer.
> 
>    2. Redistributions in binary form must reproduce the above copyright notice, this list
>       of conditions and the following disclaimer in the documentation and/or other materials
>       provided with the distribution.
> 
> THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDER ``AS IS'' AND ANY EXPRESS OR IMPLIED
> WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND
> FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL <COPYRIGHT HOLDER> OR
> CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
> CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
> SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
> ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
> NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
> ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// EOF README.md  (this is a markdown document and tested OK with blackfriday)
