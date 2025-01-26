package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	s := NewStack(1024)
	s.Push(1)
	s.Push(2)
	assert.Equal(t, 2, s.Pop())
	assert.Equal(t, 1, s.Pop())
	s.Push(3)
	assert.Equal(t, 3, s.Pop())
}

func TestStackstr(t *testing.T) {
	s := NewStack(100)
	s.Push(0x91)
	assert.Equal(t, 0x91, s.Pop())
}

func TestVM(t *testing.T) {
	// data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}

	data := []byte{0x46, 0x0c, 0x4f, 0x0c, 0x44, 0x0c, 0x03, 0x0a, 0x0d}
	vm := NewVM(data)
	assert.Nil(t, vm.Run())
	fmt.Println(string(vm.stack.Pop().([]byte)))
}
