package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"myblockchain/types"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

func (k PrivateKey) Sign(data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, k.key, data)
	if err != nil {
		return nil, err
	}

	return &Signature{r, s}, nil
}

func GeneratePrivateKey() PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	return PrivateKey{key: key}
}

func (k PrivateKey) PublicKey() PublicKey {
	return PublicKey{key: &k.key.PublicKey}
}

type PublicKey struct {
	key *ecdsa.PublicKey
}

// ToSlice returns the public key as a byte slice.
func (k PublicKey) ToSlice() []byte {
	return elliptic.MarshalCompressed(k.key, k.key.X, k.key.Y)
}

func (k PublicKey) Address() types.Address {
	h := sha256.Sum256(k.ToSlice())

	return types.AddressFromBytes(h[len(h)-20:])
}

type Signature struct {
	R, S *big.Int
}

func (s Signature) Verify(data []byte, pubKey PublicKey) bool {
	return ecdsa.Verify(pubKey.key, data, s.R, s.S)
}
