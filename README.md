# fswatch
**fswatch** is a command tool. Use file system event to trigger user defined commands. 

I reviewed the first version of fswatch(which was taged 0.1). The code I look now is shit. So I deleted almost 80% code, And left some very useful functions.

This version is works fine on mac and linux. Sorry for windows.

## How to use
**step 1.** first you need to install it by the following command.

	go get github.com/shxsun/fswatch

**step 2.** create a `.fswatch.json` file, which can be created by run 

	fswatch

modify `.fswatch.json`

**step 3.** call `fswatch` again.

## For easy coding
just execute under go test dir, that was so simple

	fswatch go test -v

Chinese Blog: <http://my.oschina.net/goskyblue/blog/194240>

## Friendly link: 
* [bee](https://github.com/astaxie/bee)
* [fsnotify](https://github.com/howeyc/fsnotify)
