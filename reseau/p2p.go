package reseau

import (
	"blockchain-m2isd/blockchain"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/vrecan/death/v3"
)

const (
	protocol   = "tcp"
	version    = 1
	commandLen = 12
)

var (
	nodeAdr         string
	minerAdr        string
	KnownNodes      = []string{"host.docker.internal:3000"}
	blocksInTransit = [][]byte{}
	memoryPool      = make(map[string]blockchain.Transaction)
)

type Addr struct {
	AddrList []string
}

type Block struct {
	AddrFrom string
	Block    []byte
}

type GetBlocks struct {
	AddrFrom string
}

type GetData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type Inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type Tx struct {
	AddrFrom    string
	Transaction []byte
}

type Version struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

func CmdToBytes(cmd string) []byte {
	var bytes [commandLen]byte

	for i, c := range cmd {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func BytesToCmd(bytes []byte) string {
	var cmd []byte

	for _, b := range bytes {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}

	return fmt.Sprintf("%s", cmd)
}

func ExtractCmd(request []byte) []byte {
	return request[:commandLen]
}

func SendAddr(adr string) {
	nodes := Addr{KnownNodes}

	nodes.AddrList = append(nodes.AddrList, nodeAdr)
	payload := GobEncode(nodes)
	req := append(CmdToBytes("addr"), payload...)
	SendData(adr, req)
}

func SendBlock(addr string, b *blockchain.Block) {
	data := Block{nodeAdr, b.Serialize()}
	payload := GobEncode(data)
	req := append(CmdToBytes("block"), payload...)
	SendData(addr, req)
}

func SendInv(adr, kind string, items [][]byte) {
	inventory := Inv{nodeAdr, kind, items}
	payload := GobEncode(inventory)
	req := append(CmdToBytes("inv"), payload...)
	SendData(adr, req)

}

func SendTx(addr string, tnx *blockchain.Transaction) {
	data := Tx{nodeAdr, tnx.Serialize()}
	payload := GobEncode(data)
	req := append(CmdToBytes("tx"), payload...)
	SendData(addr, req)
}

func SendVersion(addr string, chain *blockchain.BlockChain) {

	bestH := chain.GetBestHeight()

	payload := GobEncode(Version{version, bestH, nodeAdr})
	req := append(CmdToBytes("version"), payload...)
	SendData(addr, req)
}

func SendData(addr string, data []byte) {
	send_adr := "host.docker.internal:" + strings.Split(addr, ":")[1]
	fmt.Printf(" addr %s \n", addr)
	fmt.Printf(" send_adr %s\n", send_adr)

	conn, err := net.Dial(protocol, send_adr)

	if err != nil {
		fmt.Printf("L'Adr %s n'est pas valide \n", addr)
		var updatedNodes []string

		for _, node := range KnownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		KnownNodes = updatedNodes

		return
	}

	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))

	if err != nil {
		log.Panic(err)
	}

}

func HandleConnection(conn net.Conn, chain *blockchain.BlockChain) {
	req, err := ioutil.ReadAll(conn)
	defer conn.Close()

	if err != nil {
		log.Panic(err)
	}

	commandd := BytesToCmd(req[:commandLen])
	fmt.Printf("Reception de la commande - %s -\n", commandd)

	switch commandd {
	case "addr":
		HandleAddr(req)
	case "block":
		HandleBlock(req, chain)
	case "inv":
		HandleInv(req, chain)
	case "getblocks":
		HandleGetBlocks(req, chain)
	case "getdata":
		HandleGetData(req, chain)
	case "tx":
		HandleTx(req, chain)
	case "version":
		HandleVersion(req, chain)
	default:
		fmt.Println("Commande inconnue")
	}
}

func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func CloseDB(chain *blockchain.BlockChain) {
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)
		defer runtime.Goexit()
		chain.Database.Close()
	})
}

func SendGetBlocks(adr string) {
	payload := GobEncode(GetBlocks{nodeAdr})
	req := append(CmdToBytes("getblocks"), payload...)
	SendData(adr, req)
}

func SendGetData(adr, kind string, id []byte) {
	payload := GobEncode(GetData{nodeAdr, kind, id})
	req := append(CmdToBytes("getdata"), payload...)
	SendData(adr, req)
}

func HandleAddr(req []byte) {

	var buff bytes.Buffer
	var payload Addr

	buff.Write(req[commandLen:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	KnownNodes = append(KnownNodes, payload.AddrList...)
	fmt.Printf("Nombre de node connus %d", len(KnownNodes))
	RequestBlocks()
}

func HandleGetBlocks(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload GetBlocks

	buff.Write(request[commandLen:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)

	if err != nil {
		log.Panic(err)
	}

	blocks := chain.GetBlockHashes()

	SendInv(payload.AddrFrom, "block", blocks)
}

func HandleGetData(req []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload GetData

	buff.Write(req[commandLen:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := chain.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}
		SendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := memoryPool[txID]
		SendTx(payload.AddrFrom, &tx)
	}
}

func HandleVersion(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Version

	buff.Write(request[commandLen:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)

	if err != nil {
		log.Panic(err)
	}

	bestHeight := chain.GetBestHeight()
	otherHeight := payload.BestHeight

	if bestHeight < otherHeight {
		SendGetBlocks(payload.AddrFrom)
	} else if bestHeight > otherHeight {
		SendVersion(payload.AddrFrom, chain)
	}

	if !NodeIsKnow(payload.AddrFrom) {
		KnownNodes = append(KnownNodes, payload.AddrFrom)
	}

}

func NodeIsKnow(addr string) bool {
	for _, node := range KnownNodes {
		if node == addr {
			return true
		}
	}
	return false
}

func HandleTx(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Tx

	buff.Write(request[commandLen:])

	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)

	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := blockchain.DeserializeTransaction(txData)
	memoryPool[hex.EncodeToString(tx.ID)] = tx

	fmt.Printf("%s , %d", nodeAdr, len(memoryPool))

	if nodeAdr == KnownNodes[0] {

		for _, node := range KnownNodes {
			if node != nodeAdr && node != payload.AddrFrom {
				SendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(memoryPool) >= 2 && len(minerAdr) > 0 {
			MineTx(chain)
		}
	}

}

func MineTx(chain *blockchain.BlockChain) {
	var txs []*blockchain.Transaction

	for id := range memoryPool {
		fmt.Printf("tx: %s\n", memoryPool[id].ID)
		tx := memoryPool[id]
		if chain.VerifyTransaction(&tx) {
			txs = append(txs, &tx)
		}
	}

	if len(txs) == 0 {
		fmt.Println("Toutes les  Transactions sont invalide")
		return
	}

	cbTx := blockchain.CoinbaseTx(minerAdr, "")
	txs = append(txs, cbTx)

	newBlock := chain.MineBlock(txs)
	UTXOSet := blockchain.UTXOSet{chain}
	UTXOSet.Reindex()

	fmt.Println("Nouveau Block miné")

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		delete(memoryPool, txID)
	}

	for _, node := range KnownNodes {
		if node != nodeAdr {
			SendInv(node, "block", [][]byte{newBlock.Hash})
		}
	}

	if len(memoryPool) > 0 {
		MineTx(chain)
	}
}

func HandleInv(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Inv

	buff.Write(request[commandLen:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Reception d'une commande inv avec %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		SendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if memoryPool[hex.EncodeToString(txID)].ID == nil {
			SendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func HandleBlock(req []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Block

	buff.Write(req[commandLen:])

	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := blockchain.Deserialize(blockData)
	fmt.Println("Reception d'une nouveau block")
	chain.AddBlock(block)
	fmt.Printf("Block ajouté %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		SendGetData(payload.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]

	} else {
		UTXOSet := blockchain.UTXOSet{chain}
		UTXOSet.Reindex()
	}
}

func RequestBlocks() {
	for _, node := range KnownNodes {
		SendGetBlocks(node)
	}
}

func StartServer(nodeID, adr string) {
	fmt.Println(adr)
	nodeAdr = fmt.Sprintf("localhost:%s", nodeID)
	minerAdr = adr
	fmt.Printf("adr : %s\n", nodeAdr)

	ln, err := net.Listen(protocol, nodeAdr)

	if err != nil {
		log.Panic(err)
	}

	defer ln.Close()

	chain := blockchain.ContinueBlockChain(nodeID)

	defer chain.Database.Close()
	go CloseDB(chain)

	if nodeAdr != KnownNodes[0] {

		SendVersion(KnownNodes[0], chain)
	}
	for {
		conn, err := ln.Accept()
		fmt.Println("NOUVELLE CONNECTION")
		if err != nil {
			log.Panic(err)
		}

		go HandleConnection(conn, chain)

	}
}
