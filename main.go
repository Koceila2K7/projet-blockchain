package main

import (
	"blockchain-m2isd/commandline"
	"os"
)

func main() {
	defer os.Exit(0)
	cli := commandline.CommandLine{}
	cli.Run()

}
