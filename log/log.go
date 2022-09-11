package log

import "fmt"

func Debugf(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	fmt.Println(s)
}

func Errorf(format string, a ...any) {

}

func Warnf(format string, a ...any) {
}
