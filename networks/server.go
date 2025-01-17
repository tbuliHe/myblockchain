package networks

import (
	"fmt"
	"myblockchain/core"
	"myblockchain/crypto"
	"time"

	"github.com/sirupsen/logrus"
)

var defaultBlockTime = 5 * time.Second

type ServerOptions struct {
	RPCHandler RPCHandler
	Transport  []Transport
	BlockTime  time.Duration
	PrivateKey *crypto.PrivateKey
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
	s := &Server{
		ServerOptions: opts,
		blockTime:     opts.BlockTime,
		memPool:       NewTxPool(),
		isValidator:   opts.PrivateKey != nil,
		rpcch:         make(chan RPC),
		quitch:        make(chan struct{}, 1),
	}
	if opts.RPCHandler == nil {
		opts.RPCHandler = NewDefaultRPCHandler(s)
	}
	s.ServerOptions = opts
	return s
}

func (s *Server) Start() {
	s.initTransports()
	ticker := time.NewTicker(s.blockTime)
free:
	for {
		select {
		case rpc := <-s.rpcch:
			// handle the rpc
			if err := s.RPCHandler.HandleRPC(rpc); err != nil {
				logrus.Error(err)
			}
			fmt.Println("Received RPC", rpc)
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

func (s *Server) ProcessTransaction(from NetAddr, tx *core.Transaction) error {
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
	return s.memPool.Add(tx)
}

func (s *Server) createNewBlock() error {
	fmt.Println("Creating a new block")
	return nil
}

func (s *Server) initTransports() {
	for _, t := range s.Transport {
		go func(t Transport) {
			for rpc := range t.Consumer() {
				// handle the rpc
				s.rpcch <- rpc
			}
		}(t)
	}
}
