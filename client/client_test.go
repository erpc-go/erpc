package client

import (
	"net/http"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestCall(t *testing.T) {

	c := http.Client{}
	r, _ := http.NewRequest("", "", nil)
	c.Do(r)
}

func TestIsIp(t *testing.T) {
	c := NewClient()
	assert.Equal(t, c.isIp("ip://127.0.0.1"), true)
	assert.Equal(t, c.isIp("ip://127.0.0"), false)
}
