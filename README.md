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

## Server Design Architecture

The package indexer (server) is required to handle multiple simultaneous client
connections, each performing package index operations; possibly in parallel. In
designing this project, I considered these three architectural patterns.

### 1. Goroutine-per-Connection with Synchronized Shared State

#### Pattern

The server accepts all incoming connections, and spawns a Goroutine per client.
Each client processes one command at a time, sends a response, and loops until
the client disconnects. A shared, in-memory package dependency graph is
protected via synchronization constructs (e.g. mutexes).

#### Advantages

This approach is idiomatic and simple in Golang; making it easy to understand
and implement. It should scale well for moderate to high levels of concurrency,
and as long as the data structure is fast with insertions and deletions, there
won't be any significant contention issues.

#### Disadvantages

Locking must be performed with care to prevent race conditions or deadlocks.
And, if the load becomes too high, the mutex can become a bottleneck due to very
high contention.

### 2. Dedicated State Manager Goroutine

#### Pattern

Effectively the actor model, here, all client-handling Goroutines send commands
to a dedicated state manager Goroutine via channels. This single Goroutine
serializes all access to the package dependency graph and processes commands in
order, sending responses back through channels.

#### Advantages

This approach completely eliminates the need for synchronization constructs like
mutexes. It thereby avoids potential lock contention and data races. It can
simplify reasoning about state changes since all modifications occur in a single
thread of execution.

#### Disadvantages

The architecture is more complex to implement and maintain, as it involves
message passing and coordination between Goroutines.

### 3. Copy-on-Write

#### Pattern

State updates are done by creating new copies of the data structures, which are
swapped in atomically. Readers can access the previous snapshot without locking,
while writers build new versions.

#### Advantages

The main advantage is lock-free reads, improving read scalability.

#### Disadvantages

The overhead of cloning the entire package dependency graph on each update can
be expensive in terms of both memory and CPU usage. It is generally
inappropriate for this kind of write-heavy, low-latency scenario this server
requires.

### Summary Table

| **Pattern**                      | **Simplicity** | **Idiomatic** | **Scalability** | **Downsides**                             |
| -------------------------------- | :------------: | :-----------: | :-------------: | ----------------------------------------- |
| Goroutine-per-Connection + Mutex |      High      |      Yes      |      High       | Requires careful synchronization          |
| Dedicated State Goroutine        |     Medium     |      Yes      |     Medium      | Centralized bottleneck + added complexity |
| Copy-on-Write Immutable State    |      Low       |      No       |   Medium-High   | High overhead + complex                   |

### → Chosen Design Architecture

The final implementation uses the Goroutine-per-Connection with Mutex-Protected
Shared State. I think this approach strikes a good balance between simplicity,
performance, and idiomatic Go design. This approach passes the high-concurrency
stress tests without excessive complexity or resource consumption.

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
