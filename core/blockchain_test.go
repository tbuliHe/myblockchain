package core

import (
	"fmt"
	"myblockchain/types"
	"os"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
)

func NewBlockChainWithGenesis(t *testing.T) *BlockChain {
	l := log.NewLogfmtLogger(os.Stderr)
	bc, err := NewBlockChain(l, randBlock(t, 0, types.Hash{}))
	assert.Nil(t, err)
	return bc
}

func TestAddBlock(t *testing.T) {
	bc := NewBlockChainWithGenesis(t)
	for i := 0; i < 100; i++ {
		b := randBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		err := bc.AddBlock(b)
		assert.Nil(t, err)
	}
	assert.Equal(t, bc.Height(), uint32(100))
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

func TestGetHeader(t *testing.T) {
	bc := NewBlockChainWithGenesis(t)
	for i := 0; i < 100; i++ {
		b := randBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(b))
		header, err := bc.GetHeader(b.Height)
		assert.Nil(t, err)
		assert.Equal(t, header, b.Header)
	}
}

func TestAddBlockToHigh(t *testing.T) {
	bc := NewBlockChainWithGenesis(t)
	assert.Nil(t, bc.AddBlock(randBlock(t, 1, getPrevBlockHash(t, bc, uint32(1)))))
	assert.NotNil(t, bc.AddBlock(randBlock(t, 3, types.Hash{})))

}

func getPrevBlockHash(t *testing.T, bc *BlockChain, height uint32) types.Hash {
	prevheader, err := bc.GetHeader(height - 1)
	assert.Nil(t, err)
	return BlockHasher{}.Hash(prevheader)
}

func TestGetBlock(t *testing.T) {
	bc := NewBlockChainWithGenesis(t)
	lenBlocks := 100
	for i := 0; i < lenBlocks; i++ {
		block := randBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
		fetchedBlock, err := bc.GetBlock(block.Height)
		assert.Nil(t, err)
		assert.Equal(t, fetchedBlock, block)
	}
}
