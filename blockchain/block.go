package blockchain

import (
	"bytes"
	"encoding/gob"
	"runtime"
	"time"
)

type Block struct {
	Timestamp    int64
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
	Height       int
}

// Création du block
func CreateBlock(trs []*Transaction, prevHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), []byte{}, trs, prevHash, 0, height}
	pw := Proof(block)
	nonce, hash := pw.Work()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}

	tree := NewMerkleTree(txHashes)
	return tree.RootNode.Data
}

func CreateGenesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{}, 0)
}

// Le but de cette fonction est de facilité le stockage des block dans
// Une base de données Key Value
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)
	GestionDErreur(err)
	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var b Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&b)
	GestionDErreur(err)
	return &b
}

func GestionDErreur(err error) {
	if err != nil {
		runtime.Goexit()
	}
}
