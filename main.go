package main

import (
	"time"

	"myblockchain/networks"
)

func main() {
	trLocal := networks.NewLocalTransport("LOCAL")
	trRemote := networks.NewLocalTransport("REMOTE")

	trLocal.Connect(trRemote)
	trRemote.Connect(trLocal)

	go func() {
		for {
			trRemote.SendMessage("LOCAL", []byte("Hello, World!"))
			time.Sleep(1 * time.Second)
		}
	}()
	opt := networks.ServerOptions{
		Transport: []networks.Transport{trLocal},
	}
	s := networks.NewServer(opt)
	s.Start()
}
