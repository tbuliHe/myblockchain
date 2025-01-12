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
	Height        uint64
}

type Block struct {
	Header
	Transactions []Transaction
	hash         types.Hash
	Validatar    crypto.PublicKey
	Signature    *crypto.Signature
}

func NewBlock(h *Header, txs []Transaction) *Block {
	return &Block{
		Header:       *h,
		Transactions: txs,
	}
}

func (b *Block) Sign(priv crypto.PrivateKey) error {
	sig, err := priv.Sign(b.HeaderData())
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
	if !b.Signature.Verify(b.HeaderData(), b.Validatar) {
		return fmt.Errorf("Block signature is invalid")
	}
	return nil
}

func (b *Block) Decode(r io.Reader, dec Decoder[*Block]) error {
	return dec.Decode(r, b)
}

func (b *Block) Encode(w io.Writer, enc Encoder[*Block]) error {
	return enc.Encode(w, b)
}

func (b *Block) Hash(hasher Hasher[*Block]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b)
	}
	return b.hash
}

func (b *Block) HeaderData() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	enc.Encode(b.Header)

	return buf.Bytes()

}
