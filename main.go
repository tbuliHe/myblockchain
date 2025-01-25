package main

import (
	"bytes"
	"fmt"
	"log"
	"myblockchain/core"
	"myblockchain/crypto"
	"myblockchain/networks"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	trLocal := networks.NewLocalTransport("LOCAL")
	trRemoteA := networks.NewLocalTransport("REMOTE_A")
	trRemoteB := networks.NewLocalTransport("REMOTE_B")
	trRemoteC := networks.NewLocalTransport("REMOTE_C")

	trLocal.Connect(trRemoteA)
	trLocal.Connect(trRemoteB)
	trLocal.Connect(trRemoteC)
	trRemoteA.Connect(trLocal)

	initRemoteServers([]networks.Transport{trRemoteA, trRemoteB, trRemoteC})
	// trRemote.Connect(trLocal)

	go func() {
		for {
			if err := sendTransaction(trRemoteA, trLocal.Addr()); err != nil {
				logrus.Error(err)
			}
			time.Sleep(2 * time.Second)
		}
	}()

	go func() {
		time.Sleep(7 * time.Second)
		trLate := networks.NewLocalTransport("LATE_REMOTE")
		trRemoteC.Connect(trLate)
		lateServer := makeServer(string(trLate.Addr()), trLate, nil)

		go lateServer.Start()
	}()

	privKey := crypto.GeneratePrivateKey()
	localServer := makeServer("LOCAL", trLocal, &privKey)
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
