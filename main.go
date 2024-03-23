package main

import (
	"fmt"
	"os"
	"github.com/pred695/golang-blockchain/Cli"
	"rsc.io/quote"
)

func main() {
	fmt.Println(quote.Go())
	defer os.Exit(0)
	Cli := Cli.CommandLine{}
	Cli.Run()
}
