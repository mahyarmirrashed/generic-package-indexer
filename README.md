# Generic Package Indexer

Simple, generic package indexer with optional ability to check for dependency
cycles.

## Usage

To run the package indexer, run these following commands:

```sh
docker build -t generic-package-indexer -f Containerfile .
docker run -p 8080:8080 -d generic-package-indexer
```

## Testing

To run tests, initialize your environment (a `flake.nix` is provided), then run
the following command:

```sh
go test -v ./...
```

## Design Discussions

### Server Design Architecture Discusssion

The package indexer (server) is required to handle multiple simultaneous client
connections, each performing package index operations; possibly in parallel. In
designing this project, I considered these three architectural patterns.

#### 1. Goroutine-per-Connection with Synchronized Shared State

##### Pattern

The server accepts all incoming connections, and spawns a Goroutine per client.
Each client processes one command at a time, sends a response, and loops until
the client disconnects. A shared, in-memory package dependency graph is
protected via synchronization constructs (e.g. mutexes).

##### Advantages

This approach is idiomatic and simple in Golang; making it easy to understand
and implement. It should scale well for moderate to high levels of concurrency,
and as long as the data structure is fast with insertions and deletions, there
won't be any significant contention issues.

##### Disadvantages

Locking must be performed with care to prevent race conditions or deadlocks.
And, if the load becomes too high, the mutex can become a bottleneck due to very
high contention.

#### 2. Dedicated State Manager Goroutine

##### Pattern

Effectively the actor model, here, all client-handling Goroutines send commands
to a dedicated state manager Goroutine via channels. This single Goroutine
serializes all access to the package dependency graph and processes commands in
order, sending responses back through channels.

##### Advantages

This approach completely eliminates the need for synchronization constructs like
mutexes. It thereby avoids potential lock contention and data races. It can
simplify reasoning about state changes since all modifications occur in a single
thread of execution.

##### Disadvantages

The architecture is more complex to implement and maintain, as it involves
message passing and coordination between Goroutines.

#### 3. Copy-on-Write

##### Pattern

State updates are done by creating new copies of the data structures, which are
swapped in atomically. Readers can access the previous snapshot without locking,
while writers build new versions.

##### Advantages

The main advantage is lock-free reads, improving read scalability.

##### Disadvantages

The overhead of cloning the entire package dependency graph on each update can
be expensive in terms of both memory and CPU usage. It is generally
inappropriate for this kind of write-heavy, low-latency scenario this server
requires.

#### Summary Table

| **Pattern**                      | **Simplicity** | **Idiomatic** | **Scalability** | **Downsides**                             |
| -------------------------------- | :------------: | :-----------: | :-------------: | ----------------------------------------- |
| Goroutine-per-Connection + Mutex |      High      |      Yes      |      High       | Requires careful synchronization          |
| Dedicated State Goroutine        |     Medium     |      Yes      |     Medium      | Centralized bottleneck + added complexity |
| Copy-on-Write Immutable State    |      Low       |      No       |   Medium-High   | High overhead + complex                   |

#### → Chosen Design Architecture

The final implementation uses the Goroutine-per-Connection with Mutex-Protected
Shared State. I think this approach strikes a good balance between simplicity,
performance, and idiomatic Go design. This approach passes the high-concurrency
stress tests without excessive complexity or resource consumption.

### Dependency Graph Representation Discussion

The package dependency graph is the core data structure that the server will
manage. I considered a couple of approaches for representing this graph in the
indexer.

#### Dependency Graph with Nodes (Unidirectional Map)

##### Pattern

Create a graph data structure. Every package is mapped to a node. When a new
package is indexed, look inside the map for all needed dependencies, then add
edges to show dependency relations.

##### Advantages

First approach that I considered. Minimal memory overhead since only
dependencies are stored.

##### Disadvantages

Removing a package requires scanning the entire graph to find all packages that
depend on it. This becomes really inefficient for large graphs.

#### Forward and Reverse Dependency Hashmaps (Bi-directional Maps)

##### Pattern

Evolving from the dependency graph pattern, there's no real advantage to having
actual nodes. We can reduce to a map of maps. Then, to resolve the issue with
reverse dependencies, we can create an additional map of maps to track the
reverse dependencies as well.

##### Advantages

The main advantage is a simpler codebase and efficient lookups for dependency
validation and removal safety checks.

##### Disadvantages

Slightly more complex to implement due to maintaining two synchronized maps.

#### Adjacency Matrix Representation

##### Pattern

Using adjacency matrices, we can likely implement built-in graph traversal,
cycle detection, and other graph algorithms as needed.

##### Advantages

Using adjacency matrices, we could probably get built-in graph traversal, cycle
detection, and other graph algorithms if we need them.

##### Disadvantages

External dependencies are explicitly not allowed. Not to mention, adjacency
matrices are notoriously space-inefficient for sparse dependency graphs, which
are typical in package indexing. And, if we did decide to use a "sparse"
adjacency matrix, that's essentially our previous map-of-maps approach.

#### Summary Table

| Pattern             | Simplicity | Memory Efficiency | Lookup Performance           | Removal Efficiency           | Implementation Complexity    |
| ------------------- | :--------: | :---------------: | ---------------------------- | ---------------------------- | ---------------------------- |
| Unidirectional Map  |    High    |       High        | O(1) for dependencies        | Very Poor (full scan needed) | Medium                       |
| Bi-directional Maps |   Medium   |      Medium       | O(1) for deps and dependents | Excellent (direct lookup)    | High (two maps to sync)      |
| Adjacency Matrix    |    Low     |        Low        | O(1) fixed-size matrix       | Good (indices direct)        | High (space + external deps) |

#### → Chosen Design Architecture

The final implementation uses the bi-directional maps approach. This data
structure strikes a good balance between simplicity, performance, and
concurrency support. It allows quick validation of dependencies, safe removal
checks, and efficient cycle detection.

## Usage of LLMs

I used a conservative amount of AI-generated code in this exercise.

### Indexer

I asked Perplexity's models to review my `indexer.go` file to make sure that I
was not missing any edge cases. It correctly identified that my package indexer
was not checking for cyclic dependencies when indexing a new package. This could
create a deadlock where those packages could then never subsequently be removed.
I then asked it to generate the relevant `hasCycle` method and associated tests.

### Connection Handler

After writing out the test cases for the package indexer and the command parser,
I was pretty tired. Out of laziness, I gave Perplexity's models my connection
handler code and asked it to generate the tests. Honestly, I wasn't really sure
how to fake the I/O connections, so that probably played a substantial part in
that decision.
