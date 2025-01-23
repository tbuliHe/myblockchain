package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"myblockchain/crypto"
	"myblockchain/types"
	"time"
)

type Header struct {
	Version       uint32
	DataHash      types.Hash
	PrevBlockHash types.Hash
	Timestamp     uint64
	Height        uint32
}

func (h *Header) Bytes() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	enc.Encode(h)
	return buf.Bytes()
}

type Block struct {
	*Header
	Transactions []*Transaction
	hash         types.Hash
	Validatar    crypto.PublicKey
	Signature    *crypto.Signature
}

func NewBlock(h *Header, txs []*Transaction) *Block {
	return &Block{
		Header:       h,
		Transactions: txs,
	}
}

func NewBlockFromHeader(prev *Header, txx []*Transaction) (*Block, error) {
	dataHash, err := CaculateDataHash(txx)
	if err != nil {
		return nil, err
	}
	header := &Header{
		Version:       1,
		Height:        prev.Height + 1,
		DataHash:      dataHash,
		PrevBlockHash: BlockHasher{}.Hash(prev),
		Timestamp:     uint64(time.Now().UnixNano()),
	}
	return NewBlock(header, txx), nil

}

func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

func (b *Block) Sign(priv crypto.PrivateKey) error {
	sig, err := priv.Sign(b.Header.Bytes())
	if err != nil {
		return err
	}
	b.Validatar = priv.PublicKey()
	b.Signature = sig
	return nil
}

func (b *Block) Verify() error {
	if b.Signature == nil {
		return fmt.Errorf("Block is not signed")
	}
	if !b.Signature.Verify(b.Header.Bytes(), b.Validatar) {
		return fmt.Errorf("Block signature is invalid")
	}
	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}

	dataHash, err := CaculateDataHash(b.Transactions)
	if err != nil {
		return err
	}
	if dataHash != b.DataHash {
		return fmt.Errorf("block (%d) data hash is invalid", b.Height)
	}

	return nil
}

func (b *Block) Decode(dec Decoder[*Block]) error {
	return dec.Decode(b)
}

func (b *Block) Encode(enc Encoder[*Block]) error {
	return enc.Encode(b)
}

func (b *Block) Hash(hasher Hasher[*Header]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b.Header)
	}
	return b.hash
}

func CaculateDataHash(txs []*Transaction) (hash types.Hash, err error) {
	buf := &bytes.Buffer{}

	for _, tx := range txs {
		if err := tx.Encode(NewGobTxEncoder(buf)); err != nil {
			return types.Hash{}, err
		}
	}
	hash = sha256.Sum256(buf.Bytes())
	return hash, nil

}
