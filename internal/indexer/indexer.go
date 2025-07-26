package indexer

import "sync"

type Indexer struct {
	mu                  sync.RWMutex
	indexedPackages     map[string]struct{}
	dependencies        map[string]map[string]struct{}
	reverseDependencies map[string]map[string]struct{}
}

func New() *Indexer {
	return &Indexer{
		indexedPackages:     make(map[string]struct{}),
		dependencies:        make(map[string]map[string]struct{}),
		reverseDependencies: make(map[string]map[string]struct{}),
	}
}

func (idx *Indexer) Count() int {
	// Acquire lock
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return len(idx.indexedPackages)
}

func (idx *Indexer) Index(pkg string, deps []string) bool {
	// Acquire lock
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Check dependencies
	for _, d := range deps {
		if _, ok := idx.indexedPackages[d]; !ok {
			return false // Dependency is missing
		}
	}

	// TODO: dependency cycle detection

	// Update old reverse dependencies (if they exist)
	previousDeps, ok := idx.dependencies[pkg]
	if ok {
		for d := range previousDeps {
			revDeps := idx.reverseDependencies[d]
			if revDeps != nil {
				delete(revDeps, pkg)
				if len(revDeps) == 0 {
					delete(idx.reverseDependencies, d)
				}
			}
		}
	}

	// Add/update reverse dependencies
	depMapping := make(map[string]struct{}, len(deps))
	for _, d := range deps {
		depMapping[d] = struct{}{}

		if idx.reverseDependencies[d] == nil {
			idx.reverseDependencies[d] = make(map[string]struct{})
		}
		idx.reverseDependencies[d][pkg] = struct{}{}
	}
	idx.dependencies[pkg] = depMapping
	idx.indexedPackages[pkg] = struct{}{}

	return true
}

func (idx *Indexer) Remove(pkg string) bool {
	// Acquire lock
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Return true if package not indexed (idempotent operation)
	if _, ok := idx.indexedPackages[pkg]; !ok {
		return true
	}

	// Check if dependency of other packages
	revDeps, ok := idx.reverseDependencies[pkg]
	if ok && len(revDeps) > 0 {
		return false
	}

	// Remove the package
	delete(idx.indexedPackages, pkg)

	// Remove from reverse dependencies
	deps, ok := idx.dependencies[pkg]
	if ok {
		for d := range deps {
			revDeps := idx.reverseDependencies[d]
			if revDeps != nil {
				delete(revDeps, pkg)
				if len(revDeps) == 0 {
					delete(idx.reverseDependencies, d)
				}
			}
		}
		delete(idx.dependencies, pkg)
	}
	delete(idx.reverseDependencies, pkg)

	return true
}

func (idx *Indexer) Query(pkg string) bool {
	// Acquire lock
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	_, ok := idx.indexedPackages[pkg]
	return ok
}
