package networks

import (
	"bytes"
	"fmt"
	"myblockchain/core"
	"myblockchain/crypto"
	"time"

	"github.com/sirupsen/logrus"
)

var defaultBlockTime = 5 * time.Second

type ServerOptions struct {
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transports    []Transport
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}
type Server struct {
	ServerOptions
	blockTime   time.Duration
	memPool     *TxPool
	isValidator bool
	rpcch       chan RPC
	quitch      chan struct{}
}

func NewServer(opts ServerOptions) *Server {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}
	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}
	s := &Server{
		ServerOptions: opts,
		blockTime:     opts.BlockTime,
		memPool:       NewTxPool(),
		isValidator:   opts.PrivateKey != nil,
		rpcch:         make(chan RPC),
		quitch:        make(chan struct{}, 1),
	}
	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}
	return s
}

func (s *Server) Start() {
	s.initTransports()
	ticker := time.NewTicker(s.blockTime)
free:
	for {
		select {
		case rpc := <-s.rpcch:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				logrus.Error(err)
			}
			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				logrus.Error(err)
			}
		case <-s.quitch:
			break free
		case <-ticker.C:
			if s.isValidator {
				// create a new block
				s.createNewBlock()
			}
		default:
			// do nothing
			continue
		}
	}
	fmt.Println("Server Stopped")
}

func (s *Server) ProcessMessage(msg *DecodedMessage) error {

	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.ProcessTransaction(t)
	}
	return nil

}

func (s *Server) broadcast(payload []byte) error {
	for _, tr := range s.Transports {
		if err := tr.Broadcast(payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) ProcessTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	if s.memPool.Has(hash) {
		logrus.WithFields(logrus.Fields{
			"hash": hash,
		}).Info("tx already in mempool")
	}
	if err := tx.Verify(); err != nil {
		return err
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	logrus.WithFields(logrus.Fields{
		"hash":           hash,
		"mempool length": s.memPool.Len(),
	}).Info("adding new tx to mempool")

	go s.broadcastTransactions(tx)
	return s.memPool.Add(tx)
}

func (s *Server) broadcastTransactions(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeTx, buf.Bytes())
	return s.broadcast(msg.Bytes())
}

func (s *Server) createNewBlock() error {
	fmt.Println("Creating a new block")
	return nil
}

func (s *Server) initTransports() {
	for _, t := range s.Transports {
		go func(t Transport) {
			for rpc := range t.Consumer() {
				// handle the rpc
				s.rpcch <- rpc
			}
		}(t)
	}
}
