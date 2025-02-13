package core

import (
	"fmt"
	"myblockchain/types"
	"sync"

	"github.com/go-kit/log"
)

type BlockChain struct {
	Logger        log.Logger
	store         Storage
	lock          sync.RWMutex
	headers       []*Header
	blocks        []*Block
	txStore       map[types.Hash]*Transaction
	blockStore    map[types.Hash]*Block
	validator     Validator
	contractState *State
}

func NewBlockChain(l log.Logger, genesis *Block) (*BlockChain, error) {
	bc := &BlockChain{
		store:         NewMemoryStore(),
		headers:       []*Header{},
		Logger:        l,
		blockStore:    make(map[types.Hash]*Block),
		txStore:       make(map[types.Hash]*Transaction),
		contractState: NewState(),
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

	// for _, tx := range b.Transactions {
	// 	bc.Logger.Log("msg", "executing code", "len", len(tx.Data), "hash", tx.Hash(&TxHasher{}))
	// 	vm := NewVM(tx.Data, bc.contractState)
	// 	vm.Run()
	// 	bc.Logger.Log("msg", "executed result", "result", vm.stack.data[vm.stack.sp])
	// }
	// fmt.Printf("state => %+v\n", bc.contractState.data)

	return bc.addBlockWithoutValidation(b)
}

func (bc *BlockChain) GetBlockByHash(hash types.Hash) (*Block, error) {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	block, ok := bc.blockStore[hash]
	if !ok {
		return nil, fmt.Errorf("block with hash (%s) not found", hash)
	}
	return block, nil
}

func (bc *BlockChain) GetTxByHash(hash types.Hash) (*Transaction, error) {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	tx, ok := bc.txStore[hash]
	if !ok {
		return nil, fmt.Errorf("could not find tx with hash (%s)", hash)
	}
	return tx, nil
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
	bc.blocks = append(bc.blocks, b)
	bc.lock.Unlock()

	bc.blockStore[b.Hash(BlockHasher{})] = b
	for _, tx := range b.Transactions {
		bc.txStore[tx.Hash(TxHasher{})] = tx
	}
	return bc.store.Put(b)
}

func (bc *BlockChain) GetBlock(height uint32) (*Block, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}
	bc.lock.Lock()
	defer bc.lock.Unlock()
	return bc.blocks[height], nil
}
