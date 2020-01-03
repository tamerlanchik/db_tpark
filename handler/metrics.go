package handler

import (
	"time"
)

type Logger interface {
	Init()
	Write(path string, val time.Time)
	Data() map[string][]int64
}

type MockLogger struct {}

func (l *MockLogger) Init() {}
func (l *MockLogger) Write(path string, val time.Time) {}
func (l *MockLogger) Data() map[string][]int64 {
	return make(map[string][]int64)
}