package networks

import (
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
	assert.Equal(t, rpc.Payload, msg)
	assert.Equal(t, rpc.From, tr1.addr)
}
