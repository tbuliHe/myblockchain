package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVM(t *testing.T) {
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	vm := NewVM(data)
	assert.Nil(t, vm.Run())
	assert.Equal(t, byte(3), vm.stack[vm.sp])
	fmt.Printf("%+v", vm.stack)
}
