run : jce2go
	./jce2go -o test2go -mod github.com/edte/jce2go  test/*

debug : jce2go
	./jce2go -o test2go -mod github.com/edte/jce2go  -debug test/*

jce2go : gen_go.go lex.go main.go parse.go version.go
	go build .

build : gen_go.go lex.go main.go parse.go version.go
	go build .

update :
	go get -u
	go mod tidy

help : jce2go
	./jce2go -h

h : jce2go
	./jce2go -h

clean :
	rm -rf test2go

.PHONY : clean jce2go
