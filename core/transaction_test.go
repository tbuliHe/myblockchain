package core

import (
	"myblockchain/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignTransaction(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	data := []byte("data")
	tx := &Transaction{Data: data}

	assert.Nil(t, tx.Sign(privKey))
}

func TestVerifyTransaction(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	data := []byte("data")
	tx := &Transaction{Data: data}

	assert.Nil(t, tx.Sign(privKey))
	assert.Nil(t, tx.Verify())
	otherPrivKey := crypto.GeneratePrivateKey()
	tx.PublicKey = otherPrivKey.PublicKey()
	assert.NotNil(t, tx.Verify())
}
