package networks

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestConnect(t *testing.T) {
// 	tr1 := NewLocalTransport("A")
// 	tr2 := NewLocalTransport("B")
// 	tr1.Connect(tr2)
// 	tr2.Connect(tr1)
// 	assert.Equal(t, tr1.peers[tr2.Addr()], tr2)
// 	assert.Equal(t, tr2.peers[tr1.Addr()], tr1)
// }

// func TestSendMessage(t *testing.T) {
// 	tr1 := NewLocalTransport("A")
// 	tr2 := NewLocalTransport("B")
// 	tr1.Connect(tr2)
// 	tr2.Connect(tr1)
// 	msg := []byte("Hello, World!")
// 	assert.Nil(t, tr1.SendMessage(tr2.addr, msg))
// 	// check if the message is received
// 	rpc := <-tr2.Consumer()
// 	buf, err := io.ReadAll(rpc.Payload)
// 	assert.Nil(t, err)
// 	assert.Equal(t, buf, msg)
// 	assert.Equal(t, rpc.From, tr1.addr)
// }

// func TestBroadcast(t *testing.T) {
// 	tra := NewLocalTransport("A")
// 	trb := NewLocalTransport("B")
// 	trc := NewLocalTransport("C")
// 	tra.Connect(trb)
// 	tra.Connect(trc)
// 	msg := []byte("Hello, World!")
// 	assert.Nil(t, tra.Broadcast(msg))
// 	rbc_b := <-trb.Consumer()
// 	b, err := io.ReadAll(rbc_b.Payload)
// 	assert.Nil(t, err)
// 	assert.Equal(t, b, msg)
// 	rpc_c := <-trc.Consumer()
// 	c, err := io.ReadAll(rpc_c.Payload)
// 	assert.Nil(t, err)
// 	assert.Equal(t, c, msg)
// }

func TestTCPTransport_MultiPeerBroadcast(t *testing.T) {
	// 启动三个节点
	peerCh1 := make(chan *TCPPeer)
	tr1 := NewTCPTransport(":3002", peerCh1)
	assert.Nil(t, tr1.Start())

	peerCh2 := make(chan *TCPPeer)
	tr2 := NewTCPTransport(":3003", peerCh2)
	assert.Nil(t, tr2.Start())

	peerCh3 := make(chan *TCPPeer)
	tr3 := NewTCPTransport(":3004", peerCh3)
	assert.Nil(t, tr3.Start())

	// 建立连接
	conn1, err := net.Dial("tcp", ":3002")
	assert.Nil(t, err)
	conn2, err := net.Dial("tcp", ":3003")
	assert.Nil(t, err)

	peer1 := <-peerCh1
	peer2 := <-peerCh2

	// 测试消息广播
	testMsg := []byte("broadcast test")
	_, err = conn1.Write(testMsg)
	assert.Nil(t, err)
	_, err = conn2.Write(testMsg)
	assert.Nil(t, err)

	// 验证消息接收
	buf := make([]byte, 1024)
	n, err := peer1.conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, testMsg, buf[:n])

	n, err = peer2.conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, testMsg, buf[:n])
}
