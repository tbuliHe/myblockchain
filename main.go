package main

import (
	"bytes"
	"log"
	"myblockchain/core"
	"myblockchain/crypto"
	"myblockchain/networks"
	"net"
	"time"
)

func main() {

	privKey := crypto.GeneratePrivateKey()
	localNode := makeServer("LOCAL_NODE", &privKey, ":3000", []string{":4000"}, "")
	go localNode.Start()
	remoteNode := makeServer("REMOTE_NODE", nil, ":4000", []string{":5000"}, "")
	go remoteNode.Start()
	remoteNodeB := makeServer("REMOTE_NODE_B", nil, ":5000", nil, "")
	go remoteNodeB.Start()
	// time.Sleep(1 * time.Second)
	go func() {
		time.Sleep(6 * time.Second)

		// tcpTester()
		lateNode := makeServer("LATE_NODE", nil, ":6000", []string{":4000"}, "")
		go lateNode.Start()
	}()
	time.Sleep(1 * time.Second)
	// tcpTester()
	txSender()
	select {}
}

func makeServer(id string, privKey *crypto.PrivateKey, addr string, seedNodes []string, apiListenAddr string) *networks.Server {

	opt := networks.ServerOptions{
		APIListenAddr: apiListenAddr,
		PrivateKey:    privKey,
		ID:            id,
		SeedNodes:     seedNodes,
		ListenAddr:    addr,
	}
	s, err := networks.NewServer(opt)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func txSender() {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		panic(err)
	}
	privKey := crypto.GeneratePrivateKey()
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		panic(err)
	}
	msg := networks.NewMessage(networks.MessageTypeTx, buf.Bytes())
	_, err = conn.Write(msg.Bytes())
	if err != nil {
		panic(err)
	}
}
