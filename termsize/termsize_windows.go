package termsize

import "fmt"

func GetTerminalColumns() int {
	return 80
}

func Println(s ...interface{}) {
	fmt.Println(s...)
}
