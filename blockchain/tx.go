package blockchain

import (
	wallet "blockchain-m2isd/protfeuille"
	"bytes"
	"encoding/gob"
)

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

type TxOutputs struct {
	Outputs []TxOutput
}

func (outs TxOutputs) Serialize() []byte {
	var buffer bytes.Buffer
	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(outs)
	GestionDErreur(err)
	return buffer.Bytes()
}

func DeserializeOutputs(data []byte) TxOutputs {
	var outputs TxOutputs
	decode := gob.NewDecoder(bytes.NewReader(data))
	err := decode.Decode(&outputs)
	GestionDErreur(err)
	return outputs
}

type TxInput struct {
	ID        []byte
	Out       int
	Signature []byte
	PubKey    []byte
}

/*
func (out *TxInput) CanUnlock(data string) bool {
	//dans le can du bitcoin la vérification se fait avec un mécanisme de signature
	//cette signature prouve que nous possèdons la clé privé de l'adr publique qui reçoit
	//sans pour autant devoiler la clé privé
	return out.Sig == data
}

func (out *TxOutput) CanBeUnlocked(data string) bool {
	//dans le can du bitcoin la vérification se fait avec un mécanisme de signature
	//cette signature prouve que nous possèdons la clé privé de l'adr publique qui reçoit
	//sans pour autant devoiler la clé privé
	return out.PubKey == data
}
*/

func (in *TxInput) UsesKey(pubKeyHah []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHah) == 0
}

func (out *TxOutput) Lock(adr []byte) {
	pkh := wallet.Base58Decode(adr)
	pkh = pkh[1 : len(pkh)-4]
	out.PubKeyHash = pkh
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func NewTXOutput(value int, adr string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(adr))

	return txo
}
