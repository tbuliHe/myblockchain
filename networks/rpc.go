package networks

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"myblockchain/core"

	"github.com/sirupsen/logrus"
)

type MessageType byte

const (
	MessageTypeTx    MessageType = 0x1
	MessageTypeBlock MessageType = 0x2
)

type RPC struct {
	From    NetAddr
	Payload io.Reader
}

type Message struct {
	Header MessageType
	Data   []byte
}

func NewMessage(t MessageType, data []byte) *Message {
	return &Message{t, data}
}

func (msg *Message) Bytes() []byte {
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(msg)
	return buf.Bytes()

}

type DecodedMessage struct {
	From NetAddr
	Data any
}

type RPCDecodeFunc func(RPC) (*DecodedMessage, error)

func DefaultRPCDecodeFunc(rpc RPC) (*DecodedMessage, error) {
	msg := Message{}
	if err := gob.NewDecoder(rpc.Payload).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode message from %s:%s", rpc.From, err)
	}
	logrus.WithFields(logrus.Fields{
		"from": rpc.From,
		"type": msg.Header,
	}).Debug("received message")

	switch msg.Header {
	case MessageTypeTx:
		tx := new(core.Transaction)
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: tx,
		}, nil
	case MessageTypeBlock:
		block := new(core.Block)
		if err := block.Decode(core.NewGobBlockDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: block,
		}, nil
	default:
		return nil, fmt.Errorf("invalid message type: %d", msg.Header)
	}
}

// type RPCHandler interface {
// 	HandleRPC(rpc RPC) error
// }

// type DefaultRPCHandler struct {
// 	p RPCProcess
// }

// func NewDefaultRPCHandler(p RPCProcess) *DefaultRPCHandler {
// 	return &DefaultRPCHandler{p}
// }

// func (h *DefaultRPCHandler) HandleRPC(rpc RPC) error {

// }

type RPCProcessor interface {
	ProcessMessage(*DecodedMessage) error
}
