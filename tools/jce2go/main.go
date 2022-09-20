//
// jce2go
//

package main

import (
	"flag"
)

var (
	// 最终生成的代码根目录
	outdir string
	// 启动 debug 模式
	debug bool

	moduleCycle   bool
	moduleUpper   bool
	jsonOmitEmpty bool
)

func main() {
	flag.StringVar(&outdir, "o", "", "which dir to put generated code")
	flag.BoolVar(&debug, "debug", false, "enable debug mode")
	flag.Parse()

	for _, filename := range flag.Args() {
		gen := NewGenGo(filename, "", outdir)
		gen.Gen()
	}
}
