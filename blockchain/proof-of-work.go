package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

//Explication rapide de la preuve de travail :
/*
La preuve de travail (proof of work) est une solution utilisée dans les blockchains pour
résoudre le problème du consensus. Elle est principalement utilisée dans la cryptomonnaie Bitcoin.

La preuve de travail repose sur deux principes fondamentaux :
- le travail à effectuer doit être relativement difficile, avec une difficulté ajustable facilement,
- tandis que la vérification du travail doit être très simple et quasi instantanée.


Il existe une autre alternative appelée "proof of stake" (preuve d'enjeu).
Cette méthode est utilisée dans certaines blockchains, notamment dans la cryptomonnaie Ethereum.
*/

const Difficulty = 16

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func Proof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))
	return &ProofOfWork{b, target}
}

func (pow *ProofOfWork) InitData(nonce int) []byte {
	return bytes.Join([][]byte{pow.Block.PrevHash, pow.Block.HashTransactions(), ToHex(int64(nonce))}, []byte{})
}

func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func (pow *ProofOfWork) Work() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0
	for nonce < math.MaxInt64 {

		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash)
		fmt.Printf("\r%x", nonce)
		//Tester si le result répond à difficulté fixée
		intHash.SetBytes(hash[:])
		if intHash.Cmp(pow.Target) == -1 {
			//Car notre hash répond à la difficulté
			break
		} else {
			//Recommencer la boucle
			nonce++
		}
	}
	fmt.Println()
	return nonce, hash[:]
}

func (pow *ProofOfWork) ValidateWoork() bool {
	var ih big.Int

	d := pow.InitData(pow.Block.Nonce)
	h := sha256.Sum256(d)
	ih.SetBytes(h[:])

	return ih.Cmp(pow.Target) == -1
}
