package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewBlockChainWithGenesis(t *testing.T) *BlockChain {
	bc, err := NewBlockChain(randBlock(0))
	assert.Nil(t, err)
	return bc
}

func TestAddBlock(t *testing.T) {
	bc := NewBlockChainWithGenesis(t)
	for i := 0; i < 100; i++ {
		b := randBlockWithSignature(t, uint32(i+1))
		err := bc.AddBlock(b)
		assert.Nil(t, err)
	}
	assert.Equal(t, bc.Height(), uint32(100))
	assert.NotNil(t, bc.AddBlock(randBlock(91)))
}

func TestNewBlockChain(t *testing.T) {
	bc := NewBlockChainWithGenesis(t)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
	fmt.Println(bc.Height())
}

func TestHasBlock(t *testing.T) {
	bc := NewBlockChainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
}
