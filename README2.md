fswatch
=============
Watch file system events and then trigger some commands.

## Install
	go get -v github.com/codeskyblue/fswatch

### Simple Usage
For example if you want to call `make` when cplusplus and c file change. 

Just run:

	fswatch --ext cpp,c,h make

### Professional
1. Generate `.fswatch.yml
	use this command:

		fswatch gen
2. Use command arguments
	change watch files

		fswatch --ext js,cpp ls -l

	same as

		fswatch -w './:\.(js|cpp):3' -- ls -l

	more complex, actually it is better to modify .fswatch.yml file.

		fswatch -w './::\.(svn|git):2' \
			-w './templates:\.(go|cpp)$:\.svn' \
			-w './static:\.(js|html):\.min\.js' \
			-e 'PATH=/usr/local/bin' \
			--killgroup=true --signal KILL  \
			bash -c "ls -l"

