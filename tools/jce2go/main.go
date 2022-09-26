package main

import (
	"flag"
)

var (
	// 最终生成的代码根目录
	outdir string
	// 启动 debug 模式
	debug bool

	modulePath string

	jsonOmitEmpty bool
)

func main() {
	flag.StringVar(&outdir, "o", "", "which dir to put generated code")
	flag.BoolVar(&debug, "debug", false, "enable debug mode")
	flag.BoolVar(&jsonOmitEmpty, "json", false, "enable json tag")
	flag.StringVar(&modulePath, "mod", "", "model path")

	flag.Parse()

	for _, filename := range flag.Args() {
		// fmt.Println("--------------------")
		// fmt.Println(filename)
		// fmt.Println("--------------------")

		gen := NewGenerate(filename, modulePath, outdir)
		gen.Gen()
	}
}
