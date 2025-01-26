package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVMPushAndSub(t *testing.T) {
	// 测试压栈、减法操作
	data := []byte{
		0x08, 0x0a, // Push 8
		0x05, 0x0a, // Push 5
		0x0e, // Sub
	}
	state := NewState()
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())
	result := vm.stack.Pop().(int)
	assert.Equal(t, 3, result)
}

func TestVMPushAndPackBytes(t *testing.T) {
	// 测试压栈、打包字节数组操作
	data := []byte{
		0x41, 0x0c, // Push byte 'A'
		0x42, 0x0c, // Push byte 'B'
		0x43, 0x0c, // Push byte 'C'
		0x03, 0x0a, // Push nums of bytes
		0x0d, // Pack
	}
	state := NewState()
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())
	result := vm.stack.Pop().([]byte)
	assert.Equal(t, []byte{'A', 'B', 'C'}, result)
}

func TestVMStoreBytesToState(t *testing.T) {
	// 测试存储字节数组到状态
	data := []byte{
		0x61, 0x0c, // Push byte 'a'
		0x62, 0x0c, // Push byte 'b'
		0x63, 0x0c, // Push byte 'c'
		0x64, 0x0c, // Push byte 'd'
		0x04, 0x0a, // Push 4 (byte array size)
		0x0d,       // Pack "abcd" byte array
		0x01, 0x0a, // Push 1(int)
		0x0f, // Store
	}
	state := NewState()
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())
	value, err := vm.contractState.Get([]byte{'a', 'b', 'c', 'd'})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), deserializeInt64(value))
	value, err = vm.contractState.Get([]byte{'a', 'b', 'c', 'e'})
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(value))
}
