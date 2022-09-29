run : jce2go
	./jce2go -o demo2go -mod github.com/edte/jce2go  demo/*

debug : jce2go
	./jce2go -o demo2go -mod github.com/edte/jce2go  -debug demo/*

jce2go : generate.go lex.go main.go parse.go version.go
	go build .

build : generate.go lex.go main.go parse.go version.go
	go build .

update :
	go get -u
	go mod tidy

help : jce2go
	./jce2go -h

h : jce2go
	./jce2go -h

clean :
	rm -rf demo2go

.PHONY : clean jce2go
