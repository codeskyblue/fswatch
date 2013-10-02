# fswatch

command tools. When file change do user defined commands. 
(filter event by `.gitignore`)

Code refs from [bee](https://github.com/astaxie/bee), fsnotify

fswatch will follow 3 steps.

1. notify if file under current directory changes.
2. filter event by `.gitignore`
3. do user defined commands(passed by commands)

## preview
![fswatch](images/fswatch.png)
