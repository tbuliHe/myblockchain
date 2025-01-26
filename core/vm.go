package core

import (
	"encoding/binary"
)

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a // push data to stack
	InstrAdd      Instruction = 0x0b // add two numbers
	InstrPushByte Instruction = 0x0c // push byte to stack
	InstrPack     Instruction = 0x0d // pack n bytes to byte array
	InstrSub      Instruction = 0x0e // sub two numbers
	InstrStore    Instruction = 0x0f // store data to state
)

type Stack struct {
	data []any
	sp   int
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		sp:   -1,
	}
}

func (s *Stack) Push(data any) {
	s.sp++
	s.data[s.sp] = data
}

func (s *Stack) Pop() any {
	if s.sp == -1 {
		return nil
	}
	res := s.data[s.sp]
	s.sp--
	return res
}

type VM struct {
	data          []byte
	ip            int //instruction pointer
	stack         Stack
	contractState *State
}

func NewVM(data []byte, state *State) *VM {
	return &VM{
		data:          data,
		ip:            0,
		stack:         *NewStack(1024),
		contractState: state,
	}
}

func (vm *VM) Run() error {
	for {
		instr := vm.data[vm.ip]

		if err := vm.Execute(Instruction(instr)); err != nil {
			return err
		}
		vm.ip++
		if vm.ip > len(vm.data)-1 {
			break
		}
	}
	return nil
}

func (vm *VM) Execute(instr Instruction) error {
	s := &vm.stack
	switch instr {
	case InstrPushInt:
		s.Push(int(vm.data[vm.ip-1]))
	case InstrAdd:
		a := s.Pop().(int)
		b := s.Pop().(int)
		s.Push(a + b)
	case InstrPushByte:
		s.Push(byte(vm.data[vm.ip-1]))
	case InstrPack:
		n := s.Pop().(int)
		b := make([]byte, n)
		for i := n - 1; i >= 0; i-- { // 倒序弹出，保持字节序
			b[i] = s.Pop().(byte)
		}
		s.Push(b)
	case InstrSub:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := b - a
		vm.stack.Push(c)
	case InstrStore:
		var (
			value           = vm.stack.Pop()
			key             = vm.stack.Pop().([]byte)
			serializedValue []byte
		)

		switch v := value.(type) {
		case int:
			serializedValue = serializeInt64(int64(v))
		default:
			panic("TODO: unknown type")
		}
		vm.contractState.Put(key, serializedValue)

	}

	return nil
}

func serializeInt64(value int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	return buf
}
func deserializeInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}
