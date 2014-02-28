# fswatch
[![Build Status](https://drone.io/github.com/shxsun/fswatch/status.png)](https://drone.io/github.com/shxsun/fswatch/latest)
[![Total views](https://sourcegraph.com/api/repos/github.com/shxsun/fswatch/counters/views.png)](https://sourcegraph.com/github.com/shxsun/fswatch)

**fswatch** is a command tool. Use file system event to trigger user defined commands. 

fswatch will follow 3 steps.

1. notify if file under current directory changes.
2. filter event by `.gitignore` (if no .gitignore, the goto step 3)
3. do user defined commands(passed by commands)

## How to use
*I will show you an example*

* Demo Video <http://www.tudou.com/programs/view/NMOOE-Lj5Sc/>
* Demo video(no voice) <http://asciinema.org/a/7247>

```
go get github.com/shxsun/fswatch
# cd to a golang project
# ...
fswatch go test

# open a new shell, cd to the same place
touch test.go

# now fswatch should do some tests. (if nothing happens, tell me)
```

![fswatch](images/fswatch.png)

## Shell help
	Usage:
	  fswatch [OPTIONS] command [args...]
	
	Application Options:
	  -v, --verbose  Show verbose debug infomation
	      --delay=   Trigger event buffer time (0.5s)
	  -d, --depth=   depth of watch (3)
	  -e, --ext=     only watch specfied ext file (go,py,c,rb,cpp,cxx,h)
	  -p, --path=    watch path, support multi -p
	
	Help Options:
	  -h, --help     Show this help message

## Friendly link: 
* [bee](https://github.com/astaxie/bee)
* [fsnotify](https://github.com/howeyc/fsnotify)
