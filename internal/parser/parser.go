package parser

import (
	"errors"
	"strings"
)

type CommandType int

const (
	CommandIndex CommandType = iota
	CommandRemove
	CommandQuery
)

func (c CommandType) String() string {
	switch c {
	case CommandIndex:
		return "INDEX"
	case CommandRemove:
		return "REMOVE"
	case CommandQuery:
		return "QUERY"
	default:
		return "UNKNOWN"
	}
}

var (
	ErrInvalidFormat  = errors.New("invalid message format")
	ErrUnknownCommand = errors.New("unknown command")
)

type Message struct {
	Command      CommandType
	Package      string
	Dependencies []string // Empty slice if none
}

func Parse(line string) (*Message, error) {
	parts := strings.Split(line, "|")
	if len(parts) != 3 {
		return nil, ErrInvalidFormat
	}

	// Parse command
	var cmd CommandType
	switch parts[0] {
	case "INDEX":
		cmd = CommandIndex
	case "REMOVE":
		cmd = CommandRemove
	case "QUERY":
		cmd = CommandQuery
	default:
		return nil, ErrUnknownCommand
	}

	// Parse package
	pkg := parts[1]
	if pkg == "" {
		return nil, ErrInvalidFormat
	}

	// Parse dependencies
	var deps []string
	if parts[2] != "" {
		for d := range strings.SplitSeq(parts[2], ",") {
			dep := strings.TrimSpace(d)
			if dep != "" {
				deps = append(deps, dep)
			}
		}
	}

	return &Message{
		Command:      cmd,
		Package:      pkg,
		Dependencies: deps,
	}, nil
}
