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
	if err := b.Verify(); err != nil {
		return err
	}
	return nil
}
