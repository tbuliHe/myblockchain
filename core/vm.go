package core

import "fmt"

type Instruction byte

const (
	InstrPush Instruction = 0x0a // push data to stack
	InstrAdd  Instruction = 0x0b // add two numbers
)

type VM struct {
	data  []byte
	ip    int //instruction pointer
	stack []byte
	sp    int // stack pointer
}

func NewVM(data []byte) *VM {
	return &VM{
		data:  data,
		ip:    0,
		stack: make([]byte, 1024),
		sp:    -1,
	}
}

func (vm *VM) Run() error {
	for {
		instr := vm.data[vm.ip]
		fmt.Println(instr)

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
	switch instr {
	case InstrPush:
		vm.pushStack(vm.data[vm.ip-1])
	case InstrAdd:
		a := vm.popStack()
		b := vm.popStack()

		vm.pushStack(a + b)
	}
	return nil
}

func (vm *VM) pushStack(b byte) {
	vm.sp++
	vm.stack[vm.sp] = b
}

func (vm *VM) popStack() byte {
	b := vm.stack[vm.sp]
	vm.sp--
	return b
}
