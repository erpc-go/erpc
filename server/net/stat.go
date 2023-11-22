package net

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

// Version Going版本号
var Version = "2.1.0"

// RecvBytes 收包字节数
var RecvBytes uint64

// RecvPkgs 收包个数
var RecvPkgs uint64

// SendBytes 回包字节数
var SendBytes uint64

// SendPkgs 回包个数
var SendPkgs uint64

func stat(path string) {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
	if e != nil {
		fmt.Printf("open stat log file fail:%v\n", e)
		return
	}

	f.WriteString("RX bytes / packets  TX bytes / packets")
	f.WriteString("\n")
	f.WriteString("--------------------------------------")
	f.WriteString("\n")

	for {
		time.Sleep(time.Second)

		fileInfo, err := os.Stat(path)
		if os.IsNotExist(err) { // 统计文件被误删除，重新打开
			f.Close()
			f, e = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
			if e != nil {
				fmt.Printf("open stat log file fail:%v\n", e)
				return
			}
			f.WriteString("RX bytes / packets  TX bytes / packets")
			f.WriteString("\n")
			f.WriteString("--------------------------------------")
			f.WriteString("\n")
		} else if fileInfo.Size() > 1024*1024*1024*2 { // 大于2G重新新建文件
			f.Truncate(0)
			f.WriteString("RX bytes / packets  TX bytes / packets")
			f.WriteString("\n")
			f.WriteString("--------------------------------------")
			f.WriteString("\n")
		}

		f.WriteString(fmt.Sprintf("%d %d %d %d\n", RecvBytes, RecvPkgs, SendBytes, SendPkgs))
		atomic.StoreUint64(&RecvBytes, 0)
		atomic.StoreUint64(&RecvPkgs, 0)
		atomic.StoreUint64(&SendBytes, 0)
		atomic.StoreUint64(&SendPkgs, 0)
	}
}
