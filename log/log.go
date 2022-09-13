package log

import (
	"log"
)

// TODO: 链路日志、日志格式自定义，待做

// [level], time, info

type LogLevel int

const (
	DebugLevel LogLevel = 0
	InfoLevel  LogLevel = 1
	WarnLevel  LogLevel = 2
	ErrorLevel LogLevel = 3
	PanicLevel LogLevel = 4
	FatalLevel LogLevel = 5
)

var (
	logLevel LogLevel = ErrorLevel
	// logLevel LogLevel = DebugLevel
)

func SetLogLevel(l LogLevel) {
	logLevel = l
}

func Infof(format string, a ...any) {
	if logLevel > InfoLevel {
		return
	}
	log.Printf(format, a...)
}

func Debugf(format string, a ...any) {
	if logLevel > DebugLevel {
		return
	}
	log.Printf(format, a...)
}

func Errorf(format string, a ...any) {
	if logLevel > ErrorLevel {
		return
	}
	log.Printf(format, a...)
}

func Warnf(format string, a ...any) {
	if logLevel > WarnLevel {
		return
	}
	log.Printf(format, a...)
}

func Panicf(format string, a ...any) {
	if logLevel > PanicLevel {
		return
	}
	log.Panicf(format, a...)
}

func Fatalf(format string, a ...any) {
	if logLevel > PanicLevel {
		return
	}
	log.Fatalf(format, a...)
}
