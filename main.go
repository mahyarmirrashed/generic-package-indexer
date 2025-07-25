package main

import (
	"flag"
	"log/slog"
	"net"
	"os"

	"example.com/generic-package-indexer/internal/connhandler"
	"example.com/generic-package-indexer/internal/indexer"
)

func main() {
	addr := ":8080"

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})
	logger := slog.New(handler)

	detectCycles := flag.Bool("detect-cycles", false, "Detect dependency cycles in the indexer")
	flag.Parse()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("Failed to listen", "addr", addr, "error", err)
		os.Exit(1)
	}
	logger.Info("Listening on", "addr", addr)

	idx := indexer.New()
	idx.SetCycleDetection(*detectCycles)
	srv := connhandler.NewServer(idx, logger)

	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Warn("Failed to accept connection", "error", err)
			continue
		}

		go srv.HandleConnection(conn)
	}
}
