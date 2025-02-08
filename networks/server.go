package networks

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"myblockchain/core"
	"myblockchain/crypto"
	"myblockchain/types"
	"net"
	"os"
	"sync"
	"time"

	"github.com/go-kit/log"
)

var defaultBlockTime = 5 * time.Second

type ServerOptions struct {
	SeedNodes    []string
	ListenAddr   string
	TCPTransport *TCPTransport
	// Transport     Transport
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	// Transports    []Transport
	BlockTime  time.Duration
	PrivateKey *crypto.PrivateKey
}
type Server struct {
	ServerOptions
	TCPTransport *TCPTransport
	peerCh       chan *TCPPeer
	peerMap      map[net.Addr]*TCPPeer
	mu           sync.RWMutex
	chain        *core.BlockChain
	mempool      *TxPool
	isValidator  bool
	rpcch        chan RPC
	quitch       chan struct{}
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

	peerCh := make(chan *TCPPeer)
	tr := NewTCPTransport(opts.ListenAddr, peerCh)

	s := &Server{
		ServerOptions: opts,
		TCPTransport:  tr,
		peerCh:        peerCh,
		peerMap:       make(map[net.Addr]*TCPPeer),
		mempool:       NewTxPool(1000),
		isValidator:   opts.PrivateKey != nil,
		chain:         chain,
		rpcch:         make(chan RPC),
		quitch:        make(chan struct{}, 1),
	}

	s.TCPTransport.peerCh = peerCh
	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}
	if s.isValidator {
		go s.validatorLoop()
	}

	// for _, tr := range s.Transports {
	// 	if err := s.sendStatusMessage(tr); err != nil {
	// 		s.Logger.Log("msg", "Failed to send status message", "err", err)
	// 	}
	// }
	return s, nil
}

func (s *Server) Start() {
	s.TCPTransport.Start()
	time.Sleep(time.Second * 1)
	s.bootstrapNetwork()
	s.Logger.Log("msg", "accepting TCP connection on", "addr", s.ListenAddr, "id", s.ID)
free:
	for {
		select {
		case peer := <-s.peerCh:
			s.peerMap[peer.conn.RemoteAddr()] = peer
			go peer.readLoop(s.rpcch)
			if err := s.sendGetStatusMessage(peer); err != nil {
				s.Logger.Log("err", err)
				continue
			}
			s.Logger.Log("msg", "peer added to the server", "outgoing", peer.Outgoing, "addr", peer.conn.RemoteAddr())
		case rpc := <-s.rpcch:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("msg", "Failed to decode rpc", "err", err)
				continue
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
	case *GetStatusMessage:
		return s.processGetStatusMessage(msg.From, t)
	case *StatusMessage:
		return s.ProcessStatusMessage(msg.From, t)
	case *GetBlocksMessage:
		return s.processGetBlocksMessage(msg.From, t)
	case *BlocksMessage:
		return s.processBlocksMessage(msg.From, t)
	}
	return nil

}

func (s *Server) processBlocksMessage(from net.Addr, data *BlocksMessage) error {
	s.Logger.Log("msg", "received BLOCKS!!!!!!!!", "from", from)

	for _, block := range data.Blocks {
		fmt.Printf("BlOCK with %+v\n", block.Header)
		if err := s.chain.AddBlock(block); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) processGetBlocksMessage(from net.Addr, data *GetBlocksMessage) error {
	var (
		blocks    = []*core.Block{}
		ourHeight = s.chain.Height()
	)
	if data.To == 0 {
		for i := 0; i < int(ourHeight); i++ {
			block, err := s.chain.GetBlock(uint32(i))
			if err != nil {
				return err
			}
			blocks = append(blocks, block)
		}
	}
	fmt.Printf("%+v\n", blocks[0].Header)
	blocksMsg := &BlocksMessage{
		Blocks: blocks,
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(blocksMsg); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	msg := NewMessage(MessageTypeBlocks, buf.Bytes())
	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	}
	return peer.Send(msg.Bytes())
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

func (s *Server) processGetStatusMessage(from net.Addr, data *GetStatusMessage) error {
	s.Logger.Log("msg", "received getStatus message", "from", from)
	statusMessage := &StatusMessage{
		CurrentHeight: s.chain.Height(),
		ID:            s.ID,
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	}
	msg := NewMessage(MessageTypeStatus, buf.Bytes())
	return peer.Send(msg.Bytes())
}

func (s *Server) ProcessStatusMessage(from net.Addr, data *StatusMessage) error {
	s.Logger.Log("msg", "received STATUS message", "from", from)
	if data.CurrentHeight <= s.chain.Height() {
		s.Logger.Log("msg", "cannot sync blockHeight to low", "ourHeight", s.chain.Height(), "theirHeight", data.CurrentHeight, "addr", from)
		return nil
	}
	getBlocksMessage := &GetBlocksMessage{
		From: s.chain.Height(),
		To:   0,
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(getBlocksMessage); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())
	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	}
	return peer.Send(msg.Bytes())
}

func (s *Server) sendGetStatusMessage(peer *TCPPeer) error {
	var (
		getStatusMsg = new(GetStatusMessage)
		buf          = new(bytes.Buffer)
	)
	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())
	return peer.Send(msg.Bytes())
}

func (s *Server) broadcast(payload []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for netAddr, peer := range s.peerMap {
		if err := peer.Send(payload); err != nil {
			fmt.Printf("peer send error => addr %s [err: %s]\n", netAddr, err)
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

// func (s *Server) initTransports() {
// 	for _, t := range s.Transports {
// 		go func(t Transport) {
// 			for rpc := range t.Consumer() {
// 				// handle the rpc
// 				s.rpcch <- rpc
// 			}
// 		}(t)
// 	}
// }

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Height:    0,
		Timestamp: 000000,
	}
	return core.NewBlock(header, nil)
}

func (s *Server) bootstrapNetwork() {
	for _, addr := range s.SeedNodes {
		fmt.Println("trying to connect to ", addr)
		go func(addr string) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				fmt.Printf("could not connect to %+v\n", conn)
				return
			}
			s.peerCh <- &TCPPeer{
				conn: conn,
			}
		}(addr)
	}
}
