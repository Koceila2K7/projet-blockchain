package blockchain

import "github.com/dgraph-io/badger"

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}

	return iter
}
func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		var brutBlock []byte
		brutBlock, err = item.ValueCopy(brutBlock)
		block = Deserialize(brutBlock)
		return err
	})
	GestionDErreur(err)
	iter.CurrentHash = block.PrevHash
	return block
}
