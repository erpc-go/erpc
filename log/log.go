package log

import (
	"fmt"
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
	// logLevel LogLevel = ErrorLevel
	logLevel LogLevel = DebugLevel
)

func init() {
	// log.SetFlags(log.Lshortfile)
}

func SetLogLevel(l LogLevel) {
	logLevel = l
}

func Infof(format string, a ...any) {
	if logLevel > InfoLevel {
		return
	}
	log.Printf(fmt.Sprintf("[INFO] %s", fmt.Sprintf(format, a...)))
}

func Debugf(format string, a ...any) {
	if logLevel > DebugLevel {
		return
	}

	log.Printf(fmt.Sprintf("[DEBUG] %s", fmt.Sprintf(format, a...)))
}

func Errorf(format string, a ...any) {
	if logLevel > ErrorLevel {
		return
	}

	log.Printf(fmt.Sprintf("[ERROR] %s", fmt.Sprintf(format, a...)))
}

func Warnf(format string, a ...any) {
	if logLevel > WarnLevel {
		return
	}

	log.Printf(fmt.Sprintf("[WARN] %s", fmt.Sprintf(format, a...)))
}

func Panicf(format string, a ...any) {
	if logLevel > PanicLevel {
		return
	}

	log.Printf(fmt.Sprintf("[PANIC] %s", fmt.Sprintf(format, a...)))
}

func Fatalf(format string, a ...any) {
	if logLevel > PanicLevel {
		return
	}
	log.Printf(fmt.Sprintf("[FATAL] %s", fmt.Sprintf(format, a...)))
}
