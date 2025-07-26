package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	addr := ":8080"
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	log.Printf("Listening on %s", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("Received: %s", line)

		// Respond with ok to every message for now...
		_, err := fmt.Fprint(conn, "OK\n")
		if err != nil {
			log.Printf("Failed to send response: %v", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Connection error: %v", err)
	}
}
