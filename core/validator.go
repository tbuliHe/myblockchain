package core

import "fmt"

type Validator interface {
	ValidateBlock(b *Block) error
}

type BlockValidator struct {
	bc *BlockChain
}

func NewBlockValidator(bc *BlockChain) *BlockValidator {
	return &BlockValidator{bc: bc}
}

func (v *BlockValidator) ValidateBlock(b *Block) error {
	if v.bc.HasBlock(b.Height) {
		return fmt.Errorf("block %d already exists with hash %d", b.Height, b.Hash(BlockHasher{}))
	}
	if b.Height != v.bc.Height()+1 {
		return fmt.Errorf("block height %d is not the next block in the chain", b.Height)
	}
	prevHeader, err := v.bc.GetHeader(b.Height - 1)

	if err != nil {
		return err
	}

	hash := BlockHasher{}.Hash(prevHeader)

	if hash != b.PrevBlockHash {
		return fmt.Errorf("block %d has invalid prev block hash %d", b.Height, b.PrevBlockHash)
	}

	if err := b.Verify(); err != nil {
		return err
	}
	return nil
}
