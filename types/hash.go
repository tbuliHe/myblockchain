package types

import (
	"crypto/rand"
	"encoding/hex"
)

type Hash [32]uint8

func (h Hash) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}

func (h Hash) ToSlice() []byte {
	if len(h) != 32 {
		panic("given bytes should be 32 bytes long")
	}
	b := make([]byte, 32)
	for i, v := range h {
		b[i] = v
	}
	return b
}

func (h Hash) String() string {
	return hex.EncodeToString(h.ToSlice())
}

func HashFromBytes(b []byte) Hash {
	if len(b) != 32 {
		panic("given bytes should be 32 bytes long")
	}
	var h Hash
	copy(h[:], b)
	return h
}

func RandomBytes(size int) []byte {
	token := make([]byte, size)
	rand.Read(token)
	return token
}

func RandomHash() Hash {
	return HashFromBytes(RandomBytes(32))
}
