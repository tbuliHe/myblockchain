package main

import (
	"bytes"
	"math/rand"
	"myblockchain/core"
	"myblockchain/crypto"
	"myblockchain/networks"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	trLocal := networks.NewLocalTransport("LOCAL")
	trRemote := networks.NewLocalTransport("REMOTE")

	trLocal.Connect(trRemote)
	trRemote.Connect(trLocal)

	go func() {
		for {
			if err := sendTransaction(trRemote, trLocal.Addr()); err != nil {
				logrus.Error(err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	privKey := crypto.GeneratePrivateKey()
	opt := networks.ServerOptions{
		PrivateKey: &privKey,
		ID:         "LOCAL",
		Transports: []networks.Transport{trLocal},
	}
	s := networks.NewServer(opt)
	s.Start()
}

func sendTransaction(tr networks.Transport, to networks.NetAddr) error {
	privKey := crypto.GeneratePrivateKey()
	data := []byte(strconv.FormatInt(int64(rand.Intn(100000)), 10))
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := networks.NewMessage(networks.MessageTypeTx, buf.Bytes())
	return tr.SendMessage(to, msg.Bytes())

}
