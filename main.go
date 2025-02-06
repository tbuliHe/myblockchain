package main

import (
	"bytes"
	"fmt"
	"log"
	"myblockchain/core"
	"myblockchain/crypto"
	"myblockchain/networks"
	"time"
)

var transports = []networks.Transport{
	networks.NewLocalTransport("LOCAL"),
	// network.NewLocalTransport("REMOTE_B"),
	// network.NewLocalTransport("REMOTE_C"),
}

func main() {

	initRemoteServers(transports)
	// trRemote.Connect(trLocal)
	localNode := transports[0]
	trLate := networks.NewLocalTransport("LATE_NODE")
	go func() {
		time.Sleep(7 * time.Second)
		lateServer := makeServer(string(trLate.Addr()), trLate, nil)
		go lateServer.Start()
	}()

	privKey := crypto.GeneratePrivateKey()
	localServer := makeServer("LOCAL", localNode, &privKey)
	localServer.Start()
}

func initRemoteServers(trs []networks.Transport) {
	for i := 0; i < len(trs); i++ {
		id := fmt.Sprintf("REMOTE_%d", i)
		s := makeServer(id, trs[i], nil)
		go s.Start()
	}
}

func makeServer(id string, tr networks.Transport, privKey *crypto.PrivateKey) *networks.Server {

	opt := networks.ServerOptions{
		PrivateKey: privKey,
		ID:         id,
		Transport:  tr,
		Transports: []networks.Transport{tr},
	}
	s, err := networks.NewServer(opt)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func sendTransaction(tr networks.Transport, to networks.NetAddr) error {
	privKey := crypto.GeneratePrivateKey()
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := networks.NewMessage(networks.MessageTypeTx, buf.Bytes())
	return tr.SendMessage(to, msg.Bytes())

}
