package core

import (
	"fmt"
	"sync"

	"github.com/go-kit/log"
)

type BlockChain struct {
	Logger    log.Logger
	store     Storage
	lock      sync.RWMutex
	headers   []*Header
	validator Validator
}

func NewBlockChain(l log.Logger, genesis *Block) (*BlockChain, error) {
	bc := &BlockChain{
		store:   NewMemoryStore(),
		headers: []*Header{},
		Logger:  l,
	}
	bc.validator = NewBlockValidator(bc)
	err := bc.addBlockWithoutValidation(genesis)
	return bc, err
}

func (b *BlockChain) SetValidator(v Validator) {
	b.validator = v
}

func (bc *BlockChain) AddBlock(b *Block) error {
	//validate block
	if err := bc.validator.ValidateBlock(b); err != nil {
		return err
	}
	return bc.addBlockWithoutValidation(b)
}

func (bc *BlockChain) GetHeader(height uint32) (*Header, error) {
	bc.lock.RLock()
	if height > bc.Height() {
		return nil, fmt.Errorf("given height %d is greater than the blockchain height %d", height, bc.Height())
	}
	// whether defer
	bc.lock.RUnlock()
	return bc.headers[height], nil
}

func (bc *BlockChain) HasBlock(h uint32) bool {
	return h <= bc.Height()
}

func (bc *BlockChain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return uint32(len(bc.headers) - 1)
}

func (bc *BlockChain) addBlockWithoutValidation(b *Block) error {
	bc.lock.Lock()
	bc.headers = append(bc.headers, b.Header)
	bc.lock.Unlock()

	bc.Logger.Log(
		"msg", "block added",
		"hash", b.Hash(BlockHasher{}),
		"height", b.Height,
		"transactions", len(b.Transactions),
	)
	return bc.store.Put(b)
}
