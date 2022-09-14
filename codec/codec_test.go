package codec

import (
	"fmt"
	"testing"

	"github.com/edte/testpb2go/demo"
)

type Person struct {
	Age  int
	Name string
}

func NewPerson(name string, age int) *Person {
	return &Person{
		Age:  age,
		Name: name,
	}
}

func TestBinaMarshal(t *testing.T) {
	c := Coder(CodeTypeBinary)
	p := NewPerson("lily", 18)
	b, err := c.Marshal(p)
	fmt.Println(b, err)
}

func TestGob(t *testing.T) {
	c := Coder(CodeTypeGob)
	p := NewPerson("lily", 18)
	b, err := c.Marshal(p)
	fmt.Println(b, err)
	pp := &Person{}
	c.Unmarshal(b, pp)
	fmt.Println(pp)
}

func TestPB(t *testing.T) {
	c := Coder(CodeTypePb)
	p := demo.HelloRequest{
		Msg: "ping",
	}
	b, err := c.Marshal(&p)
	pp := demo.HelloRequest{}
	fmt.Println(b, err)
	c.Unmarshal(b, &pp)
	fmt.Println(pp)
}
