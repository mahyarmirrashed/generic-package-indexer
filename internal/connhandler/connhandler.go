package connhandler

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("[client %s] Received: %s", remoteAddr, line)

		// Respond with ok to every message for now...
		_, err := fmt.Fprint(conn, "OK\n")
		if err != nil {
			log.Printf("[client %s] Failed to send response: %v", remoteAddr, err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[client %s] Connection error: %v", remoteAddr, err)
	}
}
