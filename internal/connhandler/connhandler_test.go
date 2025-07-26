package connhandler

import (
	"bufio"
	"bytes"
	"log/slog"
	"net"
	"testing"

	"example.com/generic-package-indexer/internal/indexer"
)

var testLogger = slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

func createServerForTest() (*Server, *indexer.Indexer) {
	idx := indexer.New()
	return NewServer(idx, testLogger), idx
}

func TestIndexAndQueryHappyPath(t *testing.T) {
	server, _ := createServerForTest()
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go server.HandleConnection(serverConn)

	in := bufio.NewWriter(clientConn)
	out := bufio.NewReader(clientConn)

	_, _ = in.WriteString("INDEX|X|\n")
	_ = in.Flush()
	resp, _ := out.ReadString('\n')
	if resp != "OK\n" {
		t.Errorf("INDEX got %q, want OK\\n", resp)
	}

	_, _ = in.WriteString("QUERY|X|\n")
	_ = in.Flush()
	resp, _ = out.ReadString('\n')
	if resp != "OK\n" {
		t.Errorf("QUERY got %q, want OK\\n", resp)
	}
}

func TestQueryMissingPackage(t *testing.T) {
	server, _ := createServerForTest()
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()
	go server.HandleConnection(serverConn)

	in := bufio.NewWriter(clientConn)
	out := bufio.NewReader(clientConn)

	_, _ = in.WriteString("QUERY|Y|\n")
	_ = in.Flush()
	resp, _ := out.ReadString('\n')
	if resp != "FAIL\n" {
		t.Errorf("QUERY missing got %q, want FAIL\\n", resp)
	}
}

func TestInvalidCommand(t *testing.T) {
	server, _ := createServerForTest()
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()
	go server.HandleConnection(serverConn)

	in := bufio.NewWriter(clientConn)
	out := bufio.NewReader(clientConn)

	_, _ = in.WriteString("BLAH|Z|\n")
	_ = in.Flush()
	resp, _ := out.ReadString('\n')
	if resp != "ERROR\n" {
		t.Errorf("Invalid command got %q, want ERROR\\n", resp)
	}
}

func TestRemoveNonExistentPackageReturnsOk(t *testing.T) {
	server, _ := createServerForTest()
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()
	go server.HandleConnection(serverConn)

	in := bufio.NewWriter(clientConn)
	out := bufio.NewReader(clientConn)

	_, _ = in.WriteString("REMOVE|ghost|\n")
	_ = in.Flush()
	resp, _ := out.ReadString('\n')
	if resp != "OK\n" {
		t.Errorf("REMOVE non-existent got %q, want OK\\n", resp)
	}
}

func TestIndexMissingDepsReturnsFail(t *testing.T) {
	server, _ := createServerForTest()
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()
	go server.HandleConnection(serverConn)

	in := bufio.NewWriter(clientConn)
	out := bufio.NewReader(clientConn)

	_, _ = in.WriteString("INDEX|foo|bar\n")
	_ = in.Flush()
	resp, _ := out.ReadString('\n')
	if resp != "FAIL\n" {
		t.Errorf("INDEX missing dep got %q, want FAIL\\n", resp)
	}
}
