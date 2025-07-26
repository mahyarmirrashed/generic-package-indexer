package connhandler

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"example.com/generic-package-indexer/internal/parser"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("[client %s] Received: %s", remoteAddr, line)

		req, err := parser.Parse(line)
		if err != nil {
			log.Printf("[client %s] Failed to parse message: %v", remoteAddr, err)
			return
		}

		// Log parsed message for now...
		log.Printf("[client %s] Parsed message: %s %s %s", remoteAddr, req.Command, req.Package, req.Dependencies)

		// Respond with ok to every message for now...
		_, err = fmt.Fprint(conn, "OK\n")
		if err != nil {
			log.Printf("[client %s] Failed to send response: %v", remoteAddr, err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[client %s] Connection error: %v", remoteAddr, err)
	}
}
