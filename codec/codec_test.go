package codec

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

func equal(p1, p2 *Person) bool {
	return p1.Age == p2.Age && p2.Name == p1.Name
}

func TestCodec(t *testing.T) {
	codecs := []Codec{
		NewBinaryCoder(), NewGobCoder(), NewJceCoder(), NewJsonCoder(), NewMsgpackCoder(), NewPbCoder(), NewThriftCoder(), NewRawCoder(),
	}

	f := func(c Codec) {
		want := NewPerson("lily", 18)
		b, err := c.Marshal(want)
		if err != nil {
			panic(err)
		}
		got := &Person{}

		if err = c.Unmarshal(b, got); err != nil {
			panic(err)
		}

		if !equal(want, got) {
			panic(fmt.Errorf("got %v, but want %v", want, got))
		}
	}

	for _, c := range codecs {
		f(c)
	}

}
