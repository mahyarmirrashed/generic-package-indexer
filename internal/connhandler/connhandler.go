package connhandler

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"example.com/generic-package-indexer/internal/indexer"
	"example.com/generic-package-indexer/internal/parser"
)

const (
	CommandResponseOk    = "OK\n"
	CommandResponseFail  = "FAIL\n"
	CommandResponseError = "ERROR\n"
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
			if err := send(conn, CommandResponseError); err != nil {
				return // Send failure, close connection
			}
			continue // Keep processing next lines
		}

		// Log parsed message for now...
		log.Printf("[client %s] Parsed message: %s %s %s", remoteAddr, req.Command, req.Package, req.Dependencies)

		var resp string
		switch req.Command {
		case parser.CommandIndex:
			if idx.Index(req.Package, req.Dependencies) {
				resp = CommandResponseOk
			} else {
				resp = CommandResponseFail
			}
		case parser.CommandRemove:
			if idx.Remove(req.Package) {
				resp = CommandResponseOk
			} else {
				resp = CommandResponseFail
			}
		case parser.CommandQuery:
			if idx.Query(req.Package) {
				resp = CommandResponseOk
			} else {
				resp = CommandResponseFail
			}
		default:
			resp = CommandResponseError // Unknown command (should not happen)
		}

		if err := send(conn, resp); err != nil {
			return // Send failure, close connection
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[client %s] Connection error: %v", remoteAddr, err)
	}
}

func send(conn net.Conn, msg string) error {
	_, err := fmt.Fprint(conn, msg)
	if err != nil {
		log.Printf("[client %s] Failed to send response: %v", conn.RemoteAddr(), err)
	}
	return err
}
