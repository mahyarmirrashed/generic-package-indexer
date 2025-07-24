# Generic Package Indexer

## Server Design Architecture

The package indexer (server) is required to handle multiple simultaneous client connections, each performing package index operations; possibly in parallel. In designing this project, I considered these three architectural patterns.

### 1. Goroutine-per-Connection with Synchronized Shared State

#### Pattern

The server accepts all incoming connections, and spawns a Goroutine per client. Each client processes one command at a time, sends a response, and loops until the client disconnects. A shared, in-memory package dependency graph is protected via synchronization constructs (e.g. mutexes).

#### Advantages

This approach is idiomatic and simple in Golang; making it easy to understand and implement. It should scale well for moderate to high levels of concurrency, and as long as the data structure is fast with insertions and deletions, there won't be any significant contention issues.

#### Disadvantages

Locking must be performed with care to prevent race conditions or deadlocks. And, if the load becomes too high, the mutex can become a bottleneck due to very high contention.

### 2. Dedicated State Manager Goroutine

#### Pattern

Effectively the actor model, here, all client-handling Goroutines send commands to a dedicated state manager Goroutine via channels. This single Goroutine serializes all access to the package dependency graph and processes commands in order, sending responses back through channels.

#### Advantages

This approach completely eliminates the need for synchronization constructs like mutexes. It thereby avoids potential lock contention and data races. It can simplify reasoning about state changes since all modifications occur in a single thread of execution.

#### Disadvantages

The architecture is more complex to implement and maintain, as it involves message passing and coordination between Goroutines.

### 3. Copy-on-Write

#### Pattern

State updates are done by creating new copies of the data structures, which are swapped in atomically. Readers can access the previous snapshot without locking, while writers build new versions.

#### Advantages

The main advantage is lock-free reads, improving read scalability.

#### Disadvantages

The overhead of cloning the entire package dependency graph on each update can be expensive in terms of both memory and CPU usage. It is generally inappropriate for this kind of write-heavy, low-latency scenario this server requires.

### → Chosen Design Architecture

The final implementation uses the Goroutine-per-Connection with Mutex-Protected Shared State. I think this approach strikes a good balance between simplicity, performance, and idiomatic Go design. This approach passes the high-concurrency stress tests without excessive complexity or resource consumption.
