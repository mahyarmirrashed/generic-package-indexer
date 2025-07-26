package indexer

import "testing"

func TestIndexNoDependencies(t *testing.T) {
	idx := New()
	if !idx.Index("A", nil) {
		t.Fatal("Should index package with no dependencies")
	}
	if !idx.Query("A") {
		t.Fatal("Should find package after indexing")
	}
	if idx.Count() != 1 {
		t.Fatalf("Expected count 1, got %d", idx.Count())
	}
}

func TestIndexMissingDependencies(t *testing.T) {
	idx := New()
	if idx.Index("A", []string{"B"}) {
		t.Fatal("Should NOT index package if dependencies are missing")
	}
	if idx.Query("A") {
		t.Fatal("Package should NOT be indexed if dependencies missing")
	}
	if idx.Count() != 0 {
		t.Fatalf("Expected count 0, got %d", idx.Count())
	}
}

func TestIndexWithDependencies(t *testing.T) {
	idx := New()
	if !idx.Index("B", nil) {
		t.Fatal("Indexing 'B' should succeed")
	}
	if !idx.Index("A", []string{"B"}) {
		t.Fatal("Indexing 'A' with existing dependencies should succeed")
	}
	if !idx.Query("A") || !idx.Query("B") {
		t.Fatal("Should find both 'A' and 'B'")
	}
	if idx.Count() != 2 {
		t.Fatalf("Expected count 2, got %d", idx.Count())
	}
}

func TestRemovePackage(t *testing.T) {
	idx := New()
	idx.Index("A", nil)
	idx.Index("B", []string{"A"})
	// Cannot remove A because B depends on it
	if idx.Remove("A") {
		t.Fatal("Should NOT remove 'A' when 'B' depends on it")
	}
	// Remove B first, then A
	if !idx.Remove("B") {
		t.Fatal("Should remove 'B'")
	}
	if !idx.Remove("A") {
		t.Fatal("Should remove 'A'")
	}
	if idx.Count() != 0 {
		t.Fatalf("Expected count 0 after removal, got %d", idx.Count())
	}
}

func TestRemoveNonExistentPackage(t *testing.T) {
	idx := New()
	if !idx.Remove("ghost") {
		t.Fatal("Removing non-existent package should return true (idempotent)")
	}
}

func TestIdempotentIndexing(t *testing.T) {
	idx := New()
	if !idx.Index("pkg", nil) {
		t.Fatal("First indexing should succeed")
	}
	if !idx.Index("pkg", nil) {
		t.Fatal("Re-indexing same package with same deps should succeed")
	}
	if idx.Count() != 1 {
		t.Fatalf("Count should be 1, got %d", idx.Count())
	}
}

func TestCycleDetectionDisabledAllowsCycle(t *testing.T) {
	idx := New()
	idx.SetCycleDetection(false)
	// Index A and B with dependencies that create a cycle if detected
	if !idx.Index("A", nil) {
		t.Fatal("Should index A")
	}
	if !idx.Index("B", []string{"A"}) {
		t.Fatal("Should index B with dependency A")
	}
	// Because cycle detection is disabled, this will succeed even though it forms a cycle
	if !idx.Index("A", []string{"B"}) {
		t.Fatal("Should index A depending on B when cycle detection is disabled")
	}
}

func TestCycleDetectionEnabledDisallowsCycle(t *testing.T) {
	idx := New()
	idx.SetCycleDetection(true)
	if !idx.Index("A", nil) {
		t.Fatal("Should index A")
	}
	if !idx.Index("B", []string{"A"}) {
		t.Fatal("Should index B depending on A")
	}
	// Should reject cycle from A -> B -> A
	if idx.Index("A", []string{"B"}) {
		t.Fatal("Should NOT index A depending on B (would cause cycle)")
	}
	// Adding C -> B is fine
	if !idx.Index("C", []string{"B"}) {
		t.Fatal("Should index C depending on B")
	}
	// Should reject cycle B -> C -> B
	if idx.Index("B", []string{"C"}) {
		t.Fatal("Should NOT index B depending on C (would cause cycle)")
	}
}

func TestHasCycleDetected(t *testing.T) {
	idx := New()
	idx.SetCycleDetection(true)
	idx.Index("A", nil)
	idx.Index("B", []string{"A"})
	// Direct cycle
	if !idx.hasCycle("C", []string{"C"}) {
		t.Fatal("Direct self-cycle should be detected")
	}
	// Indirect cycle
	if !idx.hasCycle("A", []string{"B"}) {
		t.Fatal("Indirect cycle A->B->A should be detected")
	}
	// No cycle
	if idx.hasCycle("D", []string{"A"}) {
		t.Fatal("No cycle should be detected if D depends on A and no loop")
	}
}
