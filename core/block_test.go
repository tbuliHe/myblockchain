package core

import (
	"myblockchain/crypto"
	"myblockchain/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func randBlock(h uint64) *Block {
	header := &Header{
		Version:       1,
		PrevBlockHash: types.RandomHash(),
		Height:        h,
		Timestamp:     uint64(time.Now().UnixNano()),
	}
	tx := Transaction{
		Data: []byte("Hello World"),
	}
	return NewBlock(header, []Transaction{tx})
}

func TestBlockSign(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randBlock(1)
	assert.Nil(t, b.Sign(privKey))
	assert.NotNil(t, b.Signature)
}

func TestBlockVerify(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randBlock(1)
	assert.Nil(t, b.Sign(privKey))
	assert.Nil(t, b.Verify())
	otherPrivKey := crypto.GeneratePrivateKey()
	b.Validatar = otherPrivKey.PublicKey()
	assert.NotNil(t, b.Verify())
}
