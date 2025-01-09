package main

import (
	"time"

	"github.com/anthdm/projectx/network"
)

func main() {
	trLocal := network.NewLocalTransport("Local")
	trRemote := network.NewLocalTransport("Remote")

	trLocal.Connect(trRemote)
	trRemote.Connect(trLocal)

	go func() {
		for {
			trRemote.SendMessage("Local", []byte("Hello, World!"))
			time.Sleep(1 * time.Second)
		}
	}()
	opt := network.ServerOptions{
		Transports: []network.Transport{trLocal},
	}
	s := network.NewServer(opt)
	s.Start()
}
