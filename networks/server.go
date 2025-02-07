package networks

import (
	"bytes"
	"fmt"
	"myblockchain/core"
	"myblockchain/crypto"
	"myblockchain/types"
	"net"
	"os"
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
			s.Logger.Log("msg", "new peer connected", "addr", peer.conn.RemoteAddr())
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
		return nil
		// return s.ProcessGetStatusMessage(msg.From, t)
	case *StatusMessage:
		return nil
		// return s.ProcessStatusMessage(msg.From, t)
	case *GetBlocksMessage:
		return s.processGetBlocksMessage(msg.From, t)
	}
	return nil

}

func (s *Server) processGetBlocksMessage(from net.Addr, data *GetBlocksMessage) error {
	// panic("here")
	fmt.Printf("got get blocks message => %+v\n", data)
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

// func (s *Server) ProcessGetStatusMessage(from net.Addr, data *GetStatusMessage) error {
// 	fmt.Printf("=> received Getstatus msg from %s => %+v\n", from, data)
// 	statusMessage := &StatusMessage{
// 		CurrentHeight: s.chain.Height(),
// 		ID:            s.ID,
// 	}
// 	buf := new(bytes.Buffer)
// 	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
// 		return err
// 	}
// 	msg := NewMessage(MessageTypeStatus, buf.Bytes())
// 	return s.Transport.SendMessage(from, msg.Bytes())
// }

// func (s *Server) ProcessStatusMessage(from net.Addr, data *StatusMessage) error {
// 	if data.CurrentHeight <= s.chain.Height() {
// 		s.Logger.Log("msg", "cannot sync blockHeight to low", "ourHeight", s.chain.Height(), "theirHeight", data.CurrentHeight, "addr", from)
// 		return nil
// 	}
// 	// In this case we are 100% sure that the node has blocks heigher than us.
// 	getBlocksMessage := &GetBlocksMessage{
// 		From: s.chain.Height(),
// 		To:   0,
// 	}
// 	buf := new(bytes.Buffer)
// 	if err := gob.NewEncoder(buf).Encode(getBlocksMessage); err != nil {
// 		return err
// 	}
// 	msg := NewMessage(MessageTypeStatus, buf.Bytes())
// 	return s.Transport.SendMessage(from, msg.Bytes())
// }

// func (s *Server) sendStatusMessage(tr Transport) error {
// 	getStatus := new(GetStatusMessage)
// 	buf := bytes.NewBuffer([]byte{})
// 	if err := gob.NewEncoder(buf).Encode(getStatus); err != nil {
// 		return err
// 	}
// 	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())
// 	if err := s.Transport.SendMessage(tr.Addr(), msg.Bytes()); err != nil {
// 		return err
// 	}
// 	return nil
// }

func (s *Server) broadcast(payload []byte) error {
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
