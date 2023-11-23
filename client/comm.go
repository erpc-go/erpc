package client

import (
	"fmt"
	"time"
)

type ThcError struct {
	RspCode int32
	Msg     string
}

// Error
func (e ThcError) Error() string {
	return fmt.Sprintf("(code:%v, msg:%v)", e.RspCode, e.Msg)
}

// GetThcErrorFromMsg
func GetThcErrorFromMsg(ret int32, msg string) ThcError {
	return ThcError{RspCode: ret, Msg: msg}
}

// CreateFlow 随机生成序列号
func CreateFlow() uint32 {
	return uint32(time.Now().Nanosecond())
}
