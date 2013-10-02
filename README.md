# fswatch

Define `fswatch`: Command line tool. Use file system event to trigger user defined commands. 

Code refs from [bee](https://github.com/astaxie/bee), fsnotify

fswatch will follow 3 steps.

1. notify if file under current directory changes.
2. filter event by `.gitignore`
3. do user defined commands(passed by commands)

## How to use
**I just give a example**

```
go get github.com/shxsun/fswatch
# cd to a git project
# ...
fswatch go test

# open a new shell, cd to the same place
touch test.go

# now fswatch should do some tests. (if nothing happens, tell me)
```

![fswatch](images/fswatch.png)
