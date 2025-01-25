package networks

import (
	"bytes"
	"myblockchain/core"
	"myblockchain/crypto"
	"myblockchain/types"
	"os"
	"time"

	"github.com/go-kit/log"
)

var defaultBlockTime = 5 * time.Second

type ServerOptions struct {
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transports    []Transport
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}
type Server struct {
	ServerOptions
	chain       *core.BlockChain
	mempool     *TxPool
	isValidator bool
	rpcch       chan RPC
	quitch      chan struct{}
}

func NewServer(opts ServerOptions) (*Server, error) {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}
	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "ID", opts.ID)
	}
	chain, err := core.NewBlockChain(opts.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}
	s := &Server{
		ServerOptions: opts,
		mempool:       NewTxPool(1000),
		isValidator:   opts.PrivateKey != nil,
		chain:         chain,
		rpcch:         make(chan RPC),
		quitch:        make(chan struct{}, 1),
	}
	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}
	if s.isValidator {
		go s.validatorLoop()
	}
	return s, nil
}

func (s *Server) Start() {
	s.initTransports()
free:
	for {
		select {
		case rpc := <-s.rpcch:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("msg", "Failed to decode rpc", "err", err)
			}
			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				s.Logger.Log("msg", "Failed to process message", "err", err)
			}
		case <-s.quitch:
			break free
		default:
			// do nothing
			continue
		}
	}
	s.Logger.Log("msg", "Server is shutting down")
}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.ServerOptions.BlockTime)
	s.Logger.Log("msg", "Validator loop started", "blockTime", s.ServerOptions.BlockTime)
	for {
		<-ticker.C
		s.createNewBlock()
	}
}

func (s *Server) ProcessMessage(msg *DecodedMessage) error {

	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.ProcessTransaction(t)
	case *core.Block:
		return s.ProcessBlock(t)
	}
	return nil

}

func (s *Server) ProcessTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	if s.mempool.Contains(hash) {
		return nil
	}
	if err := tx.Verify(); err != nil {
		return err
	}

	// s.Logger.Log("msg", "Adding new transaction to mempool", "hash", hash, "mempool pending", s.mempool.PendingCount())

	go s.broadcastTransactions(tx)
	s.mempool.Add(tx)
	return nil
}

func (s *Server) ProcessBlock(b *core.Block) error {
	if err := b.Verify(); err != nil {
		return err
	}
	if err := s.chain.AddBlock(b); err != nil {
		return err
	}

	go s.broadcastBlock(b)

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

func (s *Server) broadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := b.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeBlock, buf.Bytes())
	return s.broadcast(msg.Bytes())
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
	curHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	txx := s.mempool.Pending()

	block, err := core.NewBlockFromHeader(curHeader, txx)
	if err != nil {
		return err
	}

	if err := block.Sign(*s.PrivateKey); err != nil {
		return err
	}

	if err := s.chain.AddBlock(block); err != nil {
		return err
	}

	s.mempool.ClearPending()

	go s.broadcastBlock(block)

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

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Height:    0,
		Timestamp: 000000,
	}
	return core.NewBlock(header, nil)
}
