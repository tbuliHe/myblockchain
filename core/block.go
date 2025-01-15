package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"myblockchain/crypto"
	"myblockchain/types"
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
	Transactions []Transaction
	hash         types.Hash
	Validatar    crypto.PublicKey
	Signature    *crypto.Signature
}

func NewBlock(h *Header, txs []Transaction) *Block {
	return &Block{
		Header:       h,
		Transactions: txs,
	}
}

func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, *tx)
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
	return nil
}

func (b *Block) Decode(r io.Reader, dec Decoder[*Block]) error {
	return dec.Decode(r, b)
}

func (b *Block) Encode(w io.Writer, enc Encoder[*Block]) error {
	return enc.Encode(w, b)
}

func (b *Block) Hash(hasher Hasher[*Header]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b.Header)
	}
	return b.hash
}
