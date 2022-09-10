package gob

import (
	"fmt"
	"testing"
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

func TestP(t *testing.T) {
	p := NewPerson("lily", 1)
	b, err := Marshal(&p)
	fmt.Println(b, err)
	pp := &Person{}
	Unmarshal(b, &pp)
	fmt.Println(pp)
}
