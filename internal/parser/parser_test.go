package parser

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseValidIndexWithNoDeps(t *testing.T) {
	msg, err := Parse("INDEX|A|")
	if err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if msg.Command != CommandIndex {
		t.Errorf("Expected CommandIndex, got %v", msg.Command)
	}
	if msg.Package != "A" {
		t.Errorf("Expected package 'A', got %q", msg.Package)
	}
	if len(msg.Dependencies) != 0 {
		t.Errorf("Expected no dependencies, got %v", msg.Dependencies)
	}
}

func TestParseValidIndexWithDeps(t *testing.T) {
	msg, err := Parse("INDEX|A|a,B,C ")
	if err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	wantDeps := []string{"a", "B", "C"}
	if !reflect.DeepEqual(msg.Dependencies, wantDeps) {
		t.Errorf("Dependencies = %v, want %v", msg.Dependencies, wantDeps)
	}
}

func TestParseTrimSpacesInDeps(t *testing.T) {
	msg, err := Parse("INDEX|A|   a  ,  b  , ,  ,c")
	if err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(msg.Dependencies, want) {
		t.Errorf("Trimmed dependencies = %v, want %v", msg.Dependencies, want)
	}
}

func TestParseRemove(t *testing.T) {
	msg, err := Parse("REMOVE|B|")
	if err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if msg.Command != CommandRemove {
		t.Errorf("Expected CommandRemove, got %v", msg.Command)
	}
	if msg.Package != "B" {
		t.Errorf("Expected package 'B', got %q", msg.Package)
	}
	if len(msg.Dependencies) != 0 {
		t.Errorf("Expected no dependencies, got %v", msg.Dependencies)
	}
}

func TestParseQuery(t *testing.T) {
	msg, err := Parse("QUERY|C|")
	if err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if msg.Command != CommandQuery {
		t.Errorf("Expected CommandQuery, got %v", msg.Command)
	}
	if msg.Package != "C" {
		t.Errorf("Expected package 'C', got %q", msg.Package)
	}
	if len(msg.Dependencies) != 0 {
		t.Errorf("Expected no dependencies, got %v", msg.Dependencies)
	}
}

func TestParseUnknownCommand(t *testing.T) {
	_, err := Parse("BLAH|A|")
	if !errors.Is(err, ErrUnknownCommand) {
		t.Fatalf("Expected ErrUnknownCommand, got %v", err)
	}
}

func TestParseInvalidFormatWithLessParts(t *testing.T) {
	_, err := Parse("INDEX|A")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("Expected ErrInvalidFormat, got %v", err)
	}
}

func TestParseInvalidFormatWithMoreParts(t *testing.T) {
	_, err := Parse("INDEX|A|a|b")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("Expected ErrInvalidFormat, got %v", err)
	}
}

func TestParseEmptyPackage(t *testing.T) {
	_, err := Parse("INDEX||")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("Expected ErrInvalidFormat for empty package, got %v", err)
	}
}

func TestParseEmptyDepsIgnoresBlanks(t *testing.T) {
	msg, err := Parse("INDEX|A|, , ,")
	if err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if len(msg.Dependencies) != 0 {
		t.Errorf("Expected no dependencies, got %v", msg.Dependencies)
	}
}
