package core

import (
	"bytes"
	"fmt"
	"myblockchain/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHeader_Encode_Decode(t *testing.T) {
	h := &Header{
		Version:       1,
		PrevBlockHash: types.RandomHash(),
		Timestamp:     uint64(time.Now().UnixNano()),
		Height:        10,
		Nonce:         989394,
	}
	buf := &bytes.Buffer{}
	assert.Nil(t, h.EncodeBinary(buf))

	hDecode := &Header{}
	assert.Nil(t, hDecode.DecodeBinary((buf)))
	assert.Equal(t, h.Version, hDecode.Version)

}

func Test_Block_Encode_Decode(t *testing.T) {
	b := &Block{
		Header: Header{
			Version:       1,
			PrevBlockHash: types.RandomHash(),
			Timestamp:     uint64(time.Now().UnixNano()),
			Height:        10,
			Nonce:         989394,
		},
		Transactions: nil,
	}
	buf := &bytes.Buffer{}
	assert.Nil(t, b.EncodeBinary(buf))

	bDecode := &Block{}
	assert.Nil(t, bDecode.DecodeBinary(buf))
	assert.Equal(t, b, bDecode)
}

func TestBlockHash(t *testing.T) {
	b := &Block{
		Header: Header{
			Version:       1,
			PrevBlockHash: types.RandomHash(),
			Timestamp:     uint64(time.Now().UnixNano()),
			Height:        10,
			Nonce:         989394,
		},
		Transactions: nil,
	}
	hash := b.Hash()
	fmt.Println(hash)
	assert.False(t, hash.IsZero())
}
