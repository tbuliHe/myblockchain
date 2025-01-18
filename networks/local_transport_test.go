package networks

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	tr1 := NewLocalTransport("A")
	tr2 := NewLocalTransport("B")
	tr1.Connect(tr2)
	tr2.Connect(tr1)
	assert.Equal(t, tr1.peers[tr2.Addr()], tr2)
	assert.Equal(t, tr2.peers[tr1.Addr()], tr1)
}

func TestSendMessage(t *testing.T) {
	tr1 := NewLocalTransport("A")
	tr2 := NewLocalTransport("B")
	tr1.Connect(tr2)
	tr2.Connect(tr1)
	msg := []byte("Hello, World!")
	assert.Nil(t, tr1.SendMessage(tr2.addr, msg))
	// check if the message is received
	rpc := <-tr2.Consumer()
	buf, err := io.ReadAll(rpc.Payload)
	assert.Nil(t, err)
	assert.Equal(t, buf, msg)
	assert.Equal(t, rpc.From, tr1.addr)
}

func TestBroadcast(t *testing.T) {
	tra := NewLocalTransport("A")
	trb := NewLocalTransport("B")
	trc := NewLocalTransport("C")
	tra.Connect(trb)
	tra.Connect(trc)
	msg := []byte("Hello, World!")
	assert.Nil(t, tra.Broadcast(msg))
	rbc_b := <-trb.Consumer()
	b, err := io.ReadAll(rbc_b.Payload)
	assert.Nil(t, err)
	assert.Equal(t, b, msg)
	rpc_c := <-trc.Consumer()
	c, err := io.ReadAll(rpc_c.Payload)
	assert.Nil(t, err)
	assert.Equal(t, c, msg)
}
