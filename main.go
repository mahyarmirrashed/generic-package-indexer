package main

import (
	"flag"
	"log/slog"
	"net"
	"os"
	"strings"

	"example.com/generic-package-indexer/internal/connhandler"
	"example.com/generic-package-indexer/internal/indexer"
)

func main() {
	port := flag.String("port", "8080", "TCP port to listen on")
	verbosity := flag.String("verbosity", "error", "Log verbosity level: debug, info, warn, error")
	detectCycles := flag.Bool("detect-cycles", false, "Detect dependency cycles in the indexer")
	flag.Parse()

	addr := ":" + *port

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLogLevel(*verbosity)})
	logger := slog.New(handler)

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

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		fallthrough
	default:
		return slog.LevelError
	}
}
