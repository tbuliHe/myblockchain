package core

import (
	"myblockchain/crypto"
	"myblockchain/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func randBlock(h uint32, PrevBlockHash types.Hash) *Block {
	header := &Header{
		Version:       1,
		PrevBlockHash: PrevBlockHash,
		Height:        h,
		Timestamp:     uint64(time.Now().UnixNano()),
	}

	return NewBlock(header, []Transaction{})
}

func randBlockWithSignature(t *testing.T, h uint32, prevBlockHash types.Hash) *Block {
	b := randBlock(h, prevBlockHash)
	privKey := crypto.GeneratePrivateKey()
	tx := randomTxWithSignature(t)
	b.AddTransaction(tx)
	assert.Nil(t, b.Sign(privKey))
	return b
}

func TestBlockSign(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randBlock(0, types.Hash{})
	assert.Nil(t, b.Sign(privKey))
	assert.NotNil(t, b.Signature)
}

func TestBlockVerify(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randBlock(0, types.Hash{})
	assert.Nil(t, b.Sign(privKey))
	assert.Nil(t, b.Verify())
	otherPrivKey := crypto.GeneratePrivateKey()
	b.Validatar = otherPrivKey.PublicKey()
	assert.NotNil(t, b.Verify())
}
