package connhandler

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"

	"example.com/generic-package-indexer/internal/indexer"
	"example.com/generic-package-indexer/internal/parser"
)

const (
	CommandResponseOk    = "OK\n"
	CommandResponseFail  = "FAIL\n"
	CommandResponseError = "ERROR\n"
)

type Server struct {
	idx    *indexer.Indexer
	logger *slog.Logger
}

func NewServer(idx *indexer.Indexer, logger *slog.Logger) *Server {
	return &Server{
		idx:    idx,
		logger: logger,
	}
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()
		s.logger.Info("Received", "client", remoteAddr, "line", line)

		req, err := parser.Parse(line)
		if err != nil {
			s.logger.Warn("Failed to parse message", "client", remoteAddr, "error", err)
			if err := s.send(conn, CommandResponseError, remoteAddr); err != nil {
				return // Send failure, close connection
			}
			continue // Keep processing next lines
		}

		s.logger.Debug("Parsed message",
			"client", remoteAddr,
			"command", req.Command,
			"package", req.Package,
			"dependencies", req.Dependencies,
		)

		var resp string
		switch req.Command {
		case parser.CommandIndex:
			if s.idx.Index(req.Package, req.Dependencies) {
				resp = CommandResponseOk
				s.logger.Info("INDEX succeeded", "client", remoteAddr, "package", req.Package, "dependencies", req.Dependencies)
			} else {
				resp = CommandResponseFail
				s.logger.Info("INDEX failed due to missing dependencies", "client", remoteAddr, "package", req.Package, "dependencies", req.Dependencies)
			}
		case parser.CommandRemove:
			if s.idx.Remove(req.Package) {
				resp = CommandResponseOk
				s.logger.Info("REMOVE succeeded", "client", remoteAddr, "package", req.Package)
			} else {
				resp = CommandResponseFail
				s.logger.Info("REMOVE failed due to dependent packages", "client", remoteAddr, "package", req.Package)
			}
		case parser.CommandQuery:
			if s.idx.Query(req.Package) {
				resp = CommandResponseOk
				s.logger.Info("QUERY found package", "client", remoteAddr, "package", req.Package)
			} else {
				resp = CommandResponseFail
				s.logger.Info("QUERY did not find package", "client", remoteAddr, "package", req.Package)
			}
		default:
			resp = CommandResponseError
			s.logger.Warn("Unknown command", "client", remoteAddr, "command", req.Command)
		}

		s.logger.Info("Indexed packages count", "count", s.idx.Count())

		if err := s.send(conn, resp, remoteAddr); err != nil {
			return
		}
	}

	if err := scanner.Err(); err != nil {
		s.logger.Warn("Connection error", "client", remoteAddr, "error", err)
	}
}

func (s *Server) send(conn net.Conn, msg string, remoteAddr string) error {
	_, err := fmt.Fprint(conn, msg)
	if err != nil {
		s.logger.Warn("Failed to send response", "client", remoteAddr, "error", err)
	}
	return err
}
