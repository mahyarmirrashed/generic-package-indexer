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

	remoteAddr := conn.RemoteAddr().String()
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
				log.Printf("[client %s] INDEX succeeded for package %q with dependencies %v", remoteAddr, req.Package, req.Dependencies)
			} else {
				resp = CommandResponseFail
				log.Printf("[client %s] INDEX failed for package %q due to missing dependencies %v", remoteAddr, req.Package, req.Dependencies)
			}
		case parser.CommandRemove:
			if idx.Remove(req.Package) {
				resp = CommandResponseOk
				log.Printf("[client %s] REMOVE succeeded for package %q", remoteAddr, req.Package)
			} else {
				resp = CommandResponseFail
				log.Printf("[client %s] REMOVE failed for package %q because other packages depend on it", remoteAddr, req.Package)
			}
		case parser.CommandQuery:
			if idx.Query(req.Package) {
				resp = CommandResponseOk
				log.Printf("[client %s] QUERY found package %q", remoteAddr, req.Package)
			} else {
				resp = CommandResponseFail
				log.Printf("[client %s] QUERY did not find package %q", remoteAddr, req.Package)
			}
		default:
			resp = CommandResponseError // Unknown command (should not happen)
			log.Printf("[client %s] Received unknown command %q", remoteAddr, req.Command)
		}

		// Show how many packages have been indexed...
		log.Printf("[server] %d packages indexed", idx.Count())

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
