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
	assert.Equal(t, c.isIp("ip://127.0.0.1"), false)
	assert.Equal(t, c.isIp("ip://127.0.0"), false)
	assert.Equal(t, c.isIp("324"), false)
	assert.Equal(t, c.isIp("127.0.0.1"), false)
	assert.Equal(t, c.isIp("127.0.0.1:8080"), false)
	assert.Equal(t, c.isIp("ip://127.0.0.1:80802"), false)
	assert.Equal(t, c.isIp("ip://127.0.0.1:abc"), false)
	assert.Equal(t, c.isIp("ip://127.0.0.1:8080"), true)
}
