package networks

import (
	"math/rand"
	"myblockchain/core"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxPool(t *testing.T) {
	p := NewTxPool()
	assert.Equal(t, 0, p.Len())
}
func TestTxPoolAdd(t *testing.T) {
	p := NewTxPool()
	tx := core.NewTransaction([]byte("Hello, world!"))
	assert.Nil(t, p.Add(tx))
	assert.Equal(t, 1, p.Len())
	_ = core.NewTransaction([]byte("Hello, wd!"))
	assert.Equal(t, p.Len(), 1)
	// p.Flush()
	// assert.Equal(t, 0, p.Len())
}

func TestSortTransactions(t *testing.T) {
	p := NewTxPool()
	txlen := 100
	for i := 0; i < txlen; i++ {
		tx := core.NewTransaction([]byte(strconv.FormatInt(int64(i), 10)))
		tx.SetFirstSeen(int64(i * rand.Intn(10000)))
		assert.Nil(t, p.Add(tx))
	}
	assert.Equal(t, txlen, p.Len())
	txx := p.Transactions()
	for i := 0; i < txlen-1; i++ {
		assert.True(t, txx[i].FirstSeen() < txx[i+1].FirstSeen())
	}

}
