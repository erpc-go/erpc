package main

import "github.com/edte/erpc/server"

func hanleHello(c *server.Context) {
	req := c.Request
	rsp := c.Response

}

func hanleEcho(c *server.Context) {

}

func main() {
	server.Handle("hello", hanleHello)
	server.Handle("echo", hanleEcho)
	server.Listen(":8080")
}
