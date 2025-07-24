package connhandler

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func HandleConnection(conn net.Conn) {
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
