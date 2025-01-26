package core

import "fmt"

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a // push data to stack
	InstrAdd      Instruction = 0x0b // add two numbers
	InstrPushByte Instruction = 0x0c // push byte to stack
	InstrPack     Instruction = 0x0d // pack n bytes to byte array
	InstrSub      Instruction = 0x0e // sub two numbers
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
	data  []byte
	ip    int //instruction pointer
	stack Stack
}

func NewVM(data []byte) *VM {
	return &VM{
		data:  data,
		ip:    0,
		stack: *NewStack(1024),
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
		fmt.Println(vm.stack)
		n := s.Pop().(int)
		b := make([]byte, n)
		for i := n - 1; i >= 0; i-- { // 倒序弹出，保持字节序
			b[i] = s.Pop().(byte)
		}
		s.Push(b)
	case InstrSub:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a - b
		vm.stack.Push(c)

	}
	return nil
}
