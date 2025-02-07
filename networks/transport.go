package networks

import "net"

type NetAddr string

//	type RPC struct {
//		// From is the address of the sender
//		From    NetAddr
//		Payload []byte
//	}
type Transport interface {
	// Consumer returns a channel that will receive RPCs
	Consumer() <-chan RPC
	Connect(Transport) error
	SendMessage(net.Addr, []byte) error
	Broadcast([]byte) error
	Addr() net.Addr
}
