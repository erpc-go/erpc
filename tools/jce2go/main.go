//
// jce2go
//

package main

import (
	"flag"
	"fmt"
)

var (
	// 最终生成的代码根目录
	gOutdir string
)

func main() {
	flag.StringVar(&gOutdir, "o", "", "which dir to put generated code")
	flag.Parse()

	for _, filename := range flag.Args() {
		fmt.Println("begin parse file: ", filename)

		gen := NewGenGo(filename, "", gOutdir)
		gen.Gen()
	}
}
