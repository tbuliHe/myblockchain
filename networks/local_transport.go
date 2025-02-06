package networks

import (
	"bytes"
	"fmt"
	"sync"
)

type LocalTransport struct {
	addr         NetAddr
	lock         sync.RWMutex
	peers        map[NetAddr]*LocalTransport
	ConsumerChan chan RPC
}

func NewLocalTransport(addr NetAddr) *LocalTransport {
	return &LocalTransport{
		addr:         addr,
		ConsumerChan: make(chan RPC, 1024),
		peers:        make(map[NetAddr]*LocalTransport),
	}
}
func (t *LocalTransport) Consumer() <-chan RPC {
	return t.ConsumerChan
}

func (t *LocalTransport) Connect(other Transport) error {
	tr := other.(*LocalTransport)
	t.lock.Lock()
	defer t.lock.Unlock()

	t.peers[other.Addr()] = tr
	return nil
}

func (t *LocalTransport) SendMessage(addr NetAddr, payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.addr == addr {
		return nil
	}
	peer, ok := t.peers[addr]
	if !ok {
		return fmt.Errorf("%s:Could not send meeage to %s", t.addr, addr)
	}
	peer.ConsumerChan <- RPC{
		From:    t.addr,
		Payload: bytes.NewReader(payload),
	}
	return nil
}

func (t *LocalTransport) Broadcast(payload []byte) error {
	for _, peer := range t.peers {
		if err := t.SendMessage(peer.Addr(), payload); err != nil {
			return err
		}
	}
	return nil
}

func (t *LocalTransport) Addr() NetAddr {
	return t.addr
}
