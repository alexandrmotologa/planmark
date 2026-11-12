# Architecture Proposal: Distributed Task Scheduler

## 1. Context and Goals
We need a local-first task coordinator for AI subagents. The current architecture suffers from unbounded goroutine growth and race conditions when multiple workers compete for write locks on SQLite databases.

## 2. Proposed Architecture

### 2.1 Storage Layer
We will use SQLite in WAL mode with connection pooling capped at one dedicated writer connection and up to 16 concurrent reader connections. All mutations are queued into an in-memory ring buffer with backpressure.

### 2.2 Worker Pool and Dispatch
Each worker subagent runs as an isolated process communicating through standard I/O pipes. The coordinator monitors process health using heartbeats sent every 500 milliseconds.

```go
type WorkerConfig struct {
    ID          string
    MaxRetries  int
    TimeoutSec  int
    Environment map[string]string
}
```

### 2.3 Failure Recovery
If a worker crashes, the supervisor restarts it with exponential backoff (starting at 100ms up to a ceiling of 5 seconds). Any unfinished tasks return to the priority queue.
