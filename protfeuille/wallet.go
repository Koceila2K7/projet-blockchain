package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func CreerUnePaireDeCle() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	//La graine
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pub
}

func MakeWallet() *Wallet {
	private, public := CreerUnePaireDeCle()
	wallet := Wallet{private, public}
	return &wallet
}

func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)
	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}

	return hasher.Sum(nil)
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:checksumLength]
}

func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)
	versionedHash := append([]byte{version}, pubHash...)
	chk := Checksum(versionedHash)

	fullHash := append(versionedHash, chk...)
	adr := Base58Encode(fullHash)

	//fmt.Printf("clé publique : %x\n", w.PublicKey)
	//fmt.Printf("hash publique : %x\n", pubHash)
	//fmt.Printf("Adr final : %x\n", adr)
	return adr
}

func ValiderUneAddress(adr string) bool {
	pubKeyHah := Base58Decode([]byte(adr))
	curentChk := pubKeyHah[len(pubKeyHah)-checksumLength:]
	version := pubKeyHah[0]
	pubKeyHah = pubKeyHah[1 : len(pubKeyHah)-checksumLength]
	targetChk := Checksum(append([]byte{version}, pubKeyHah...))
	return bytes.Compare(curentChk, targetChk) == 0

}
