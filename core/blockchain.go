package core

type BlockChain struct {
	store     Storage
	headers   []*Header
	validator Validator
}

func NewBlockChain(genesis *Block) (*BlockChain, error) {
	bc := &BlockChain{
		store:   NewMemoryStore(),
		headers: []*Header{},
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

func (bc *BlockChain) HasBlock(h uint32) bool {
	return h <= bc.Height()
}

func (bc *BlockChain) Height() uint32 {
	return uint32(len(bc.headers) - 1)
}

func (bc *BlockChain) addBlockWithoutValidation(b *Block) error {
	bc.headers = append(bc.headers, &b.Header)
	return bc.store.Put(b)
}
