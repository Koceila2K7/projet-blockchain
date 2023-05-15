package commandline

import (
	"blockchain-m2isd/blockchain"
	wallet "blockchain-m2isd/protfeuille"
	reseau "blockchain-m2isd/reseau"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct {
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Bienvenue dans la ligne de commande du Projet de Koceila Nabil et Anis\nVoici la liste des instructions disponibles :")
	fmt.Println("print - affiche tous les blocs de la chaîne")
	fmt.Println("getbalance -address ADDRESS - renvoie l'état d'un compte")
	fmt.Println("createblockchain -address ADDRESS crée une blockchain et l'argent miné pour le block genisis est en envoyé à l'adr donnée")
	fmt.Println("send -from FROM -to TO -amount AMOUNT -mine - envoi de l'argent (amount) de from à to, avec -mine miner sur ce noued")
	fmt.Println("createwallet - creation d'un nouveau wallet")
	fmt.Println("listaddresses - liste de toutes adr stockées en mémoire")
	fmt.Println("reindexutxo - Re-index la liste des UTXO")
	fmt.Print("startnode -miner ADR - Démmarer un noeud avec un id définie comme variable d'ENV  avec la var NODE_ID -miner avec minage")

}

func (cli *CommandLine) reindexUTXO(nodeID string) {
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()

	UTXOSet := blockchain.UTXOSet{chain}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransaction()
	fmt.Printf("Réindexation finie, il existe %d transactions UTXO \n", count)
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) printChain(nodeID string) {
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()

	iter := chain.Iterator()

	for {
		block := iter.Next()
		fmt.Printf("Hash du bloc précédent : %x\n", block.PrevHash)
		fmt.Printf("Données du bloc : %s\n", block.Transactions)
		fmt.Printf("Hash du bloc : %x\n", block.Hash)

		pw := blockchain.Proof(block)
		fmt.Printf("Pow:%s\n", strconv.FormatBool(pw.ValidateWoork()))
		fmt.Println()

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println()
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	var nodeID = os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env n'est pas définiée")
		runtime.Goexit()
	}

	getbalancecmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	createblockchaincmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getbalanceadr := getbalancecmd.String("address", "", "Adr du compte ciblé")
	createblockchainadr := createblockchaincmd.String("address", "", "adr qui va reçevoir l'argent du minage du genisis")
	sendFrom := sendCmd.String("from", "", "source de la transaction")
	sendTo := sendCmd.String("to", "", "destination de la transaction")
	sendAmout := sendCmd.Int("amount", 0, "cout de la transaction")
	sendMine := sendCmd.Bool("mine", false, "Miner directement")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")
	switch os.Args[1] {
	case "getbalance":
		err := getbalancecmd.Parse(os.Args[2:])
		blockchain.GestionDErreur(err)
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		blockchain.GestionDErreur(err)
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "createblockchain":
		err := createblockchaincmd.Parse(os.Args[2:])
		blockchain.GestionDErreur(err)
	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.GestionDErreur(err)

	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		blockchain.GestionDErreur(err)

	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		blockchain.GestionDErreur(err)

	case "reindexutxo":
		err := reindexCmd.Parse(os.Args[2:])
		blockchain.GestionDErreur(err)

	default:
		cli.printChain(nodeID)
		runtime.Goexit()
	}

	if getbalancecmd.Parsed() {
		if *getbalanceadr == "" {
			getbalancecmd.Usage()
			runtime.Goexit()
		}
		cli.getbalance(*getbalanceadr, nodeID)
	}
	if reindexCmd.Parsed() {
		cli.reindexUTXO(nodeID)
	}
	if createblockchaincmd.Parsed() {
		if *createblockchainadr == "" {
			createblockchaincmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createblockchainadr, nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmout, nodeID, *sendMine)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}
	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}
	if startNodeCmd.Parsed() {

		cli.StartNode(nodeID, *startNodeMiner)
	}
}
func (cli *CommandLine) StartNode(nodeID, minerAddress string) {
	fmt.Printf("Starting Node %s\n", nodeID)

	if len(minerAddress) > 0 {
		if wallet.ValiderUneAddress(minerAddress) {
			fmt.Println("Le minage est activé, l'adr de récompence est :  ", minerAddress)
		} else {
			log.Panic("adr de récompence non valide !")
		}
	}
	reseau.StartServer(nodeID, minerAddress)
}
func (cli *CommandLine) createBlockChain(adr, nodeID string) {
	if !wallet.ValiderUneAddress(adr) {
		log.Panic("L'adr n'est pas valide ......... ! ")
	}
	chain := blockchain.InitBlockChain(adr, nodeID)
	chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{chain}
	UTXOSet.Reindex()
	fmt.Println("Création finie .............")
}

func (cli *CommandLine) getbalance(address, nodeID string) {
	if !wallet.ValiderUneAddress(address) {
		log.Panic("L'adr n'est pas valide ......... ! ")
	}
	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOSet := blockchain.UTXOSet{chain}
	UTXOSet.Reindex()
	defer chain.Database.Close()
	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUnspentTransactions(pubKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Etat du compte %s :%d ", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !wallet.ValiderUneAddress(to) {
		log.Panic("L'adr source n'est pas valide ......... ! ")
	}
	if !wallet.ValiderUneAddress(from) {
		log.Panic("L'adr destination n'est pas valide ......... ! ")
	}
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{chain}
	UTXOSet.Reindex()
	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := blockchain.NewTransaction(&wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTx := blockchain.CoinbaseTx(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}
		block := chain.MineBlock(txs)
		UTXOSet.Update(block)
	} else {
		reseau.SendTx(reseau.KnownNodes[0], tx)
		fmt.Println("send tx")
	}

	fmt.Println("Opération réussie!.............")
}

func (cli *CommandLine) listAddresses(nodeID string) {
	wallets, _ := wallet.CreateWallets(nodeID)
	addresses := wallets.GetAllAddresses()
	for _, adr := range addresses {
		fmt.Println(adr)
	}
}

func (cli *CommandLine) createWallet(nodeID string) {
	wts, _ := wallet.CreateWallets(nodeID)
	adrs := wts.AddWallet()
	wts.SaveFile(nodeID)
	fmt.Printf("Nouvelle adr : %s\n", adrs)
}
