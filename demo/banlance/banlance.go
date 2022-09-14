package main

import "github.com/edte/erpc"

func main() {
	erpc.ListenBalance(":2531")
}
