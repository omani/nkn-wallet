package main

import (
	"log"

	cmd "github.com/omani/nkn-wallet/cmd/nkn-wallet/commands"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Panic: %+v", r)
		}
	}()

	cmd.Execute()
}
