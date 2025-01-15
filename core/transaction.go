package core

import (
	"fmt"
	"myblockchain/crypto"
)

type Transaction struct {
	Data      []byte
	From      crypto.PublicKey
	Signature *crypto.Signature
}

func (tx *Transaction) Sign(priv crypto.PrivateKey) error {
	sig, err := priv.Sign(tx.Data)
	if err != nil {
		return err
	}
	tx.From = priv.PublicKey()
	tx.Signature = sig
	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("Transaction is not signed")
	}
	if !tx.Signature.Verify(tx.Data, tx.From) {
		return fmt.Errorf("Transaction signature is invalid")
	}
	return nil
}
