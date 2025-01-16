package core

import (
	"bytes"
	"myblockchain/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignTransaction(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	data := []byte("Hello World")
	tx := &Transaction{Data: data}

	assert.Nil(t, tx.Sign(privKey))
}

func TestVerifyTransaction(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	data := []byte("Hello world")
	tx := &Transaction{Data: data}

	assert.Nil(t, tx.Sign(privKey))
	assert.Nil(t, tx.Verify())
	otherPrivKey := crypto.GeneratePrivateKey()
	tx.From = otherPrivKey.PublicKey()
	assert.NotNil(t, tx.Verify())
}

func TestTxEncodeDecode(t *testing.T) {
	tx := randomTxWithSignature(t)
	buf := &bytes.Buffer{}
	assert.Nil(t, tx.Encode(NewGobTxEncoder(buf)))

	txDecode := new(Transaction)
	assert.Nil(t, txDecode.Decode(NewGobTxDecoder(buf)))
	assert.Equal(t, tx, txDecode)
}

func randomTxWithSignature(t *testing.T) *Transaction {
	tx := &Transaction{Data: []byte("Hello World")}
	privKey := crypto.GeneratePrivateKey()
	assert.Nil(t, tx.Sign(privKey))
	return tx
}
