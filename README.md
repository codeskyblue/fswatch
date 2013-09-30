# trigger

cli tools. watch file change, and do use specfied commands.

** Developing ...**

Below is Chinese words
还是中文写起来更容易些。暂时先写中文了。

这个是我写的第二个库了，第一个是带一个日志模块，可以带颜色的(klog).

代码参考了各个地方gh(a git tool written by golang)，bee，klog。

下个定义： 监控文件的变化，根据`.gitignore`文件进行过滤，并执行相应的命令。

我假设你了解一点git的基础知识，知道`.gitignore`是做什么用的。

想了半天也没想出啥好名字，fsense ，有点像google的adsense。


我比较喜欢的一种用法，就是`fsense sh -c "go test -i && go install"`。
