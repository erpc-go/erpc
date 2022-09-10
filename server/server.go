package server

type Context struct {
	Request  interface{}
	Response interface{}
}

type HandlerFunc func(*Context)

var (
	handleMap = make(map[string]HandlerFunc, 200)
)

func Handle(partten string, handler HandlerFunc) {
	handleMap[partten] = handler
}

func Listen(addr string) {

}
