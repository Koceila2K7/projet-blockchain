package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dgraph-io/badger"
)

const (
	dbChemin    = "./tmp/blocks_%s"
	lastHashKey = "lastHash"
	genesisData = "Première transaction depuis genisis"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

func BlockChainAlreadyStored(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}
	return true
}
func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				log.Println("BDD ouverte")
				return db, nil
			}
			log.Println("impossible de d'ouvrir la BDD :", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}
func InitBlockChain(address, nodeId string) *BlockChain {
	var lastHash []byte
	chemin := fmt.Sprintf(dbChemin, nodeId)
	if BlockChainAlreadyStored(chemin) {
		fmt.Println("La blockchain est déja stockée")
		runtime.Goexit()
	}
	opts := badger.DefaultOptions(chemin)
	opts.Logger = nil
	db, err := openDB(chemin, opts)
	GestionDErreur(err)
	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData)
		genesis := CreateGenesis(cbtx)
		fmt.Println("Création du block Genesis .............. fait!")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		GestionDErreur(err)
		err = txn.Set([]byte(lastHashKey), genesis.Hash)
		lastHash = genesis.Hash
		return err
	})
	GestionDErreur(err)
	return &BlockChain{lastHash, db}
}

func ContinueBlockChain(nodeId string) *BlockChain {
	chemin := fmt.Sprintf(dbChemin, nodeId)

	if !BlockChainAlreadyStored(chemin) {
		fmt.Println("Aucune blockchain n'as été trouvée")
		runtime.Goexit()
	}
	var lastHash []byte

	opts := badger.DefaultOptions(chemin)
	opts.Logger = nil
	db, err := openDB(chemin, opts)
	GestionDErreur(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(lastHashKey))
		GestionDErreur(err)
		lastHash, err = item.ValueCopy(lastHash)
		return err
	})

	GestionDErreur(err)
	chain := BlockChain{lastHash, db}

	return &chain
}
func (chain *BlockChain) MineBlock(tx []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(lastHashKey))
		GestionDErreur(err)
		lastHash, err = item.ValueCopy(lastHash)
		item, err = txn.Get(lastHash)
		GestionDErreur(err)
		var lastBlockData []byte
		lastBlockData, _ = item.ValueCopy(lastBlockData)
		lastBlock := Deserialize(lastBlockData)
		lastHeight = lastBlock.Height
		return err
	})
	GestionDErreur(err)
	newBlock := CreateBlock(tx, lastHash, lastHeight+1)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		GestionDErreur(err)
		err = txn.Set([]byte(lastHashKey), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})
	GestionDErreur(err)
	return newBlock
}

func (chain *BlockChain) AddBlock(block *Block) {
	err := chain.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}

		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		GestionDErreur(err)

		item, err := txn.Get([]byte(lastHashKey))
		GestionDErreur(err)

		var lastHash []byte

		lastHash, _ = item.ValueCopy(lastHash)

		item, err = txn.Get(lastHash)
		GestionDErreur(err)
		var lastBlockData []byte

		lastBlockData, _ = item.ValueCopy(lastBlockData)

		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte(lastHashKey), block.Hash)
			GestionDErreur(err)
			chain.LastHash = block.Hash
		}

		return nil
	})
	GestionDErreur(err)
}

func (chain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Block  non trouvé")
		} else {
			var blockData []byte
			blockData, _ = item.ValueCopy(blockData)

			block = *Deserialize(blockData)
		}
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}
func (chain *BlockChain) GetBestHeight() int {
	var lastBlock Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(lastHashKey))

		GestionDErreur(err)

		var lastHash []byte
		lastHash, _ = item.ValueCopy(lastHash)

		item, err = txn.Get(lastHash)
		GestionDErreur(err)
		var lastBlockData []byte

		lastBlockData, _ = item.ValueCopy(lastBlockData)

		lastBlock = *Deserialize(lastBlockData)

		return nil
	})

	GestionDErreur(err)

	return lastBlock.Height
}
func (chain *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte

	iter := chain.Iterator()

	for {
		block := iter.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}

func (chain *BlockChain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXO
}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("La transaction n'existe pas ")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTxs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTx, err := bc.FindTransaction(in.ID)
		GestionDErreur(err)
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	tx.Sign(privKey, prevTxs)

}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTxs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTx, err := bc.FindTransaction(in.ID)
		GestionDErreur(err)
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	return tx.Verify(prevTxs)

}
