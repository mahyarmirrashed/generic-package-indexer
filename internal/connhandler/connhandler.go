package connhandler

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"example.com/generic-package-indexer/internal/indexer"
	"example.com/generic-package-indexer/internal/parser"
)

func HandleConnection(conn net.Conn, idx *indexer.Indexer) {
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

		switch req.Command {
		case parser.CommandIndex:
			if idx.Index(req.Package, req.Dependencies) {
				send(conn, "OK\n")
			} else {
				send(conn, "FAIL\n")
			}
		case parser.CommandRemove:
			if idx.Remove(req.Package) {
				send(conn, "OK\n")
			} else {
				send(conn, "FAIL\n")
			}
		case parser.CommandQuery:
			if idx.Query(req.Package) {
				send(conn, "OK\n")
			} else {
				send(conn, "FAIL\n")
			}
		default:
			send(conn, "ERROR\n") // Unknown command (should not happen)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[client %s] Connection error: %v", remoteAddr, err)
	}
}

func send(conn net.Conn, msg string) {
	_, err := fmt.Fprint(conn, msg)
	if err != nil {
		log.Printf("[client %s] Failed to send response: %v", conn.RemoteAddr(), err)
		return
	}
}
