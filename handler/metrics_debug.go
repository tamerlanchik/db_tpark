// +build debug

package handler

import (
	"sync"
	"time"
)

func init() {
	timeLogger = &TimeLogger{}
	timeLogger.Init()
}

type TimeLogger struct{
	data map[string][]int64
	mutex sync.Mutex
}

func (l *TimeLogger) Write(path string, val time.Time) {
	t := int64(time.Since(val)) / 1e6
	l.mutex.Lock()
	l.data[path] = append(l.data[path], t)
	l.mutex.Unlock()
}

func (l *TimeLogger) Init() {
	l.data = make(map[string][]int64)
}

func (l *TimeLogger) Data() map[string][]int64 {
	return l.data
}
