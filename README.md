# fswatch
[![Build Status](https://drone.io/github.com/shxsun/fswatch/status.png)](https://drone.io/github.com/shxsun/fswatch/latest)
[![Total views](https://sourcegraph.com/api/repos/github.com/shxsun/fswatch/counters/views.png)](https://sourcegraph.com/github.com/shxsun/fswatch)

**fswatch** is a command tool. Use file system event to trigger user defined commands. 

I reviewed the first version of fswatch(which was taged 0.1). The code I look now is shit. So I deleted almost 80% code, And left some very useful functions.

This version is works fine on mac and linux. Sorry for windows.

## How to use
first you need to install it by the following command.

	go get github.com/shxsun/fswatch

create a `.fswatch.json` file, which can be created by run `fswatch`

modify `.fswatch.json`

call `fswatch` again.

## Friendly link: 
* [bee](https://github.com/astaxie/bee)
* [fsnotify](https://github.com/howeyc/fsnotify)
