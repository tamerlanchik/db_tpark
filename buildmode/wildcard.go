package buildmode

import "fmt"

var BuildTag string

var Log Logger

type Logger struct {

}

func (l *Logger) Println(args ...interface{}){
	if BuildTag=="debug" {
		fmt.Println(args...)
	}
}
