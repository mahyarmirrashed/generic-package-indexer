package main

import (
	"log"
	"net"

	"example.com/generic-package-indexer/internal/connhandler"
	"example.com/generic-package-indexer/internal/indexer"
)

func main() {
	addr := ":8080"
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	log.Printf("Listening on %s", addr)

	idx := indexer.New()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go connhandler.HandleConnection(conn, idx)
	}
}
