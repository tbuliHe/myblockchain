package core

import (
	"myblockchain/crypto"
	"myblockchain/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func randBlock(t *testing.T, h uint32, PrevBlockHash types.Hash) *Block {
	privKey := crypto.GeneratePrivateKey()
	tx := randomTxWithSignature(t)
	header := &Header{
		Version:       1,
		PrevBlockHash: PrevBlockHash,
		Height:        h,
		Timestamp:     uint64(time.Now().UnixNano()),
	}

	b := NewBlock(header, []*Transaction{tx})
	dataHash, err := CaculateDataHash(b.Transactions)
	assert.Nil(t, err)
	b.Header.DataHash = dataHash

	assert.Nil(t, b.Sign(privKey))
	return b
}

// func randBlockWithSignature(t *testing.T, h uint32, prevBlockHash types.Hash) *Block {
// 	b := randBlock(h, prevBlockHash)

// 	tx := randomTxWithSignature(t)
// 	b.AddTransaction(tx)
// 	assert.Nil(t, b.Sign(privKey))
// 	return b
// }

func TestBlockSign(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randBlock(t, 0, types.Hash{})
	assert.Nil(t, b.Sign(privKey))
	assert.NotNil(t, b.Signature)
}

func TestBlockVerify(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randBlock(t, 0, types.Hash{})
	assert.Nil(t, b.Sign(privKey))
	assert.Nil(t, b.Verify())
	otherPrivKey := crypto.GeneratePrivateKey()
	b.Validatar = otherPrivKey.PublicKey()
	assert.NotNil(t, b.Verify())
}
