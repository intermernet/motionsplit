package main

import (
	"log"
	"os"

	"github.com/Intermernet/motionsplit"
)

func main() {
	args := os.Args
	if len(args) < 2 || len(args) > 2 {
		log.Fatal("Please provide a single filename to convert\n")
	}
	err := motionsplit.Split(args[1])
	if err != nil {
		log.Fatal(err)
	}
}
