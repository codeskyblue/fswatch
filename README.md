# fswatch
**fswatch** is a command tool. Use file system event to trigger user defined commands. 

I reviewed the first version of fswatch(which was taged 0.1). The code I look now is shit. So I deleted almost 80% code, And left some very useful functions.

This version is works fine on mac and linux. (**Support windows now**)

## Install

	go get github.com/codeskyblue/fswatch

## How to use
fswatch (default)monitor files which match regex: `\\.(go|py|php|java|cpp|h|rb)$`. This is regex can be modified through `.fswatch.json` file, we talk about that later.

This easiest way to use fswatch is run command:

	fswatch ls
	# open a new terminal (make sure the workdir some as the previos terminal)
	touch hello.go
	# some thing should happens

Now it's time to talk about the `.fswatch.json` file. The file can be created by 

	fswatch 
	# then press y when prompt occurs

modify `.fswatch.json`. and run `fswatch` again. Enjoy your time.

## Other
* fswatch kill all process when restart. (mac and linux killed by pgid, windows by taskkill)
* `Ctrl+C` will trigger fswatch quit and kill all process it started.

auto test golang code

	fswatch go test -v

Chinese Blog: <http://my.oschina.net/goskyblue/blog/194240>

## Friendly link: 
* [bee](https://github.com/astaxie/bee)
* [fsnotify](https://github.com/howeyc/fsnotify)
