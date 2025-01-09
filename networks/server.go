package networks

import (
	"fmt"
	"time"
)

type ServerOptions struct {
	Transport []Transport
}
type Server struct {
	ServerOptions
	rpcch chan RPC
	quit  chan struct{}
}

func NewServer(opts ServerOptions) *Server {
	return &Server{
		ServerOptions: opts,
		rpcch:         make(chan RPC, 1024),
		quit:          make(chan struct{}),
	}
}

func (s *Server) Start() {
	s.initTransports()
	ticker := time.NewTicker(5 * time.Second)
free:
	for {
		select {
		case rpc := <-s.rpcch:
			// handle the rpc
			fmt.Println("Received RPC", rpc)
		case <-s.quit:
			break free
		case <-ticker.C:
			fmt.Println("Server is running every 5 seconds")
		default:
			// do nothing
			continue
		}
	}
	fmt.Println("Server Stopped")
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
