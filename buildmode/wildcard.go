package buildmode

import "fmt"

var BuildTag string
var LogTag string

var Log Logger

type Logger struct {

}

func (l *Logger) Println(args ...interface{}){
	if BuildTag=="debug" && LogTag=="log"{
		fmt.Println(args...)
	}
}
