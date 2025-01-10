package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privKey := GeneratePrivateKey()
	// pubKey := privKey.PublicKey()
	// address := pubKey.Address()

	msg := []byte("Hello, world!")
	signature, err := privKey.Sign(msg)
	assert.Nil(t, err)
	assert.True(t, signature.Verify(msg, privKey.PublicKey()))

	// fmt.Println("Signature:", signature)
	// fmt.Println("Private key:", privKey)
	// fmt.Println("Public key:", pubKey)
	// fmt.Println("Address:", address)
}

func TestKeyPair_Sign_Verify(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.PublicKey()

	msg := []byte("Hello, world!")
	signature, err := privKey.Sign(msg)
	assert.Nil(t, err)

	assert.True(t, signature.Verify(msg, pubKey))
}

func TestKeyPair_Sign_Verify_Fail(t *testing.T) {
	privKey := GeneratePrivateKey()

	msg := []byte("Hello, world!")
	signature, err := privKey.Sign(msg)
	assert.Nil(t, err)
	otherPrivKey := GeneratePrivateKey()
	otherPubKey := otherPrivKey.PublicKey()
	assert.False(t, signature.Verify(msg, otherPubKey))
	assert.False(t, signature.Verify([]byte("Hello, Not world!"), privKey.PublicKey()))
}
