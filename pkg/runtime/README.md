# Runtime

The runtime is the concurrency backbone of GompherMP. It provides the primitives that the transformer-generated code calls to spawn goroutine teams, distribute work, and synchronize across goroutines. Together with the parser and the transformer, it closes the loop from a `//gompher`-annotated Go program to a correctly-running parallel binary.

Where the parser turns directives into typed structs and the transformer rewrites the AST, the runtime is what actually runs. Every `Parallel`, `For`, `Critical`, `Barrier`, etc. call in the generated code maps to a function in this package.

---

## Pipeline

```
 Transformer-generated Go code
        |
        v
 runtime.Parallel(...)
        Submits one job per worker to the persistent goroutine pool,
        registers each worker in a team context, and waits for all of them
        to finish (implicit barrier).
        |
        v
 Work-sharing constructs
        runtime.For, runtime.ForDynamic, runtime.Sections
        submit jobs to the same pool and distribute work units across the
        team according to a scheduling policy.
        |
        v
 Synchronization primitives
        runtime.Critical, runtime.Single, runtime.Master, runtime.Barrier
        coordinate goroutines inside the team (mutual exclusion, single-execution,
        master-only blocks, explicit synchronization points).
```

---

## Architecture: persistent goroutine pool

At package initialization (`init()` in `pool.go`), the runtime pre-instantiates a worker pool sized to `runtime.GOMAXPROCS(0)`. The workers are long-lived goroutines that consume jobs from a shared channel and stay alive across every parallel region in the program. This mirrors the worker-pool architecture used by production OpenMP runtimes (libgomp, Intel OpenMP) and avoids paying goroutine creation/teardown cost on every `Parallel`, `For`, `Sections` call.

### Configuration

```go
runtime.PoolSize()         // current pool size
runtime.SetPoolSize(n)     // resize the pool (creates a new pool and shuts down the old workers)
runtime.CurrentTeamSize()  // size of the team the calling goroutine belongs to (1 if none)
```

`SetPoolSize(n)` with `n <= 0` is clamped to `1` (sequential execution) so misuse degrades gracefully instead of producing a broken pool. In production, prefer setting `GOMAXPROCS` before the program starts to control the initial pool size; `SetPoolSize` is primarily a testing and tuning hook.

---

## Public API

### Work distribution - `parallel.go`

GompherMP follows the OpenMP worksharing model: a `Parallel` region creates a
team that persists for the whole region, and the worksharing constructs (`For`,
`Sections`, `ForDynamic`) distribute work **across that existing team** rather
than provisioning their own. Each is called by every team goroutine and ends at
an implicit barrier. The combined constructs (`ParallelFor`, `ParallelSections`,
`ParallelForDynamic`) are standalone sugar for a worksharing construct wrapped
in a `Parallel`.

| Function | Purpose |
|---|---|
| `Parallel(body func(int))` | Submits `PoolSize()` jobs to the pool, each receiving its thread ID, and waits for all to finish (implicit barrier). When invoked from inside an already-active parallel region, serializes the call: the body runs once with thread ID 0 in a virtual team of size 1 (nested parallelism disabled, matching OpenMP's default). |
| `For(threadID int, body func(int), iterations int)` | Worksharing static loop: the calling goroutine runs the contiguous chunk of `[0, iterations)` assigned to its `threadID` out of the current team, then waits at the implicit barrier. Standalone (no team) it runs the whole loop sequentially. |
| `ForDynamic(body func(int), iterations, chunkSize int)` | Worksharing dynamic loop: team goroutines claim chunks of `chunkSize` from the team's shared cursor until exhausted, then barrier. Suited for iterations with variable cost. Standalone it runs sequentially. |
| `Sections(sections []func())` | Worksharing: the team claims the independent blocks from the shared cursor so each runs exactly once, then barriers. Standalone it runs them sequentially. |
| `ParallelFor(body func(int), iterations int)` | Combined: `Parallel` + static `For` in one call. |
| `ParallelForDynamic(body func(int), iterations, chunkSize int)` | Combined: `Parallel` + dynamic `ForDynamic` in one call. |
| `ParallelSections(sections []func())` | Combined: `Parallel` + `Sections` in one call. |

### Pool management - `pool.go`

| Function | Purpose |
|---|---|
| `PoolSize() int` | Returns the size of the active pool. |
| `SetPoolSize(n int)` | Replaces the active pool with a new one of size `n` (clamped to 1 if non-positive). Old workers exit cleanly when their channel closes. |
| `CurrentTeamSize() int` | Returns the size of the team the calling goroutine belongs to, or `1` if the caller is not inside any parallel region. |

### Synchronization - `sync.go`

| Function | Purpose |
|---|---|
| `Critical(name string, body func())` | Mutual exclusion around `body`. An empty `name` uses a global anonymous lock; a non-empty name uses a per-name lock so different names can run in parallel. |
| `Single(body func())` | Executes `body` on exactly one goroutine of the team (elected by an atomic compare-and-swap — CAS: an atomic operation that updates a value only if it still equals an expected one — on the team token) while the others skip it, then synchronizes at the implicit barrier. Standalone it simply runs the body. |
| `Master(threadID int, body func())` | Executes `body` only when called from the master goroutine (`threadID == 0`). No implicit barrier (non-master goroutines continue immediately). |
| `Barrier()` | Synchronization point for the current team. All goroutines in the team must reach the call before any can proceed. No-op outside a parallel region. |

---

## Team semantics

A parallel region is built around a *team context* (a struct that tracks the active workers so `Barrier()` can synchronize them). When a pool worker picks up a job carrying a team context, it registers itself in the team for the duration of the job's body and unregisters when the body returns. Registration uses the worker's runtime goroutine ID (parsed from the goroutine stack header) as the key. The runtime maintains a `map[goroutineID]*teamContext` protected by an `sync.RWMutex`.

This lookup is what lets `Barrier()` know which team the caller belongs to. Calls to `Barrier()` from outside a `Parallel` region (i.e. from a goroutine without a registered team) become safe no-ops instead of panicking.

`Barrier()` is **reusable**. Once every goroutine in the team has reached it and been released, it automatically resets so it can be used again. (It is built from a `sync.Cond` plus a round counter, instead of a one-shot `sync.WaitGroup`.) This is exactly what the OpenMP worksharing model needs: each `For`, `Sections`, and `Single` ends with its own implicit barrier, so one parallel region usually hits several in a row - something the old one-shot barrier could not do.

The barrier also has a `waitThen` variant used to reset shared team state safely between constructs. The dynamic constructs (`Sections`, dynamic `For`) hand out work through a shared counter, and `Single` picks its runner through a shared flag; both must be cleared to zero before the next construct begins. `waitThen` lets the **last** goroutine to arrive run that reset while it still holds the barrier's lock - after every goroutine has finished the current construct, but before any has been woken to start the next one. That window is the only moment when no goroutine is touching the shared state, so the reset is race-free and visible to everyone.

---

## Mapping from `//gompher` directives to runtime calls

| Directive | Runtime function the transformer emits |
|---|---|
| `//gompher parallel` | `Parallel(func(threadID int) { ... })` |
| `//gompher for` (static) | `For(threadID, func(i int) { ... }, n)` |
| `//gompher for schedule(dynamic, c)` | `ForDynamic(func(i int) { ... }, n, c)` |
| `//gompher parallel for` | `ParallelFor(func(i int) { ... }, n)` |
| `//gompher parallel for schedule(dynamic, c)` | `ParallelForDynamic(func(i int) { ... }, n, c)` |
| `//gompher sections` | `Sections([]func(){ ... })` |
| `//gompher parallel sections` | `ParallelSections([]func(){ ... })` |
| `//gompher critical` | `Critical("", func() { ... })` |
| `//gompher critical(name)` | `Critical("name", func() { ... })` |
| `//gompher single` | `Single(func() { ... })` |
| `//gompher master` | `Master(threadID, func() { ... })` |
| `//gompher barrier` | `Barrier()` |
| `//gompher atomic update` | `AtomicAddInt(&x, delta)` |
| `//gompher atomic write` | `AtomicStoreInt(&x, v)` |
| `//gompher atomic read` | `v = AtomicLoadInt(&x)` |
| `//gompher atomic` | `runtime.AtomicAddInt` / `AtomicStoreInt` / `AtomicLoadInt` (lock-free helpers over `int`, see `atomic.go`) |

---

## Running Tests

The runtime ships with a full test suite covering work distribution, synchronization, edge cases, and concurrent safety.

### Run all tests

```bash
go test ./pkg/runtime/...
```

### Run with the race detector

```bash
go test -race ./pkg/runtime/...
```

The race detector catches data races between goroutines. The stress tests (`TestForDynamic_StressNoRace`, `TestCritical_PreventRaceCondition`) are designed to exercise the race detector under load.

### Generate a coverage profile

```bash
go test -coverprofile=runtime_cov.out ./pkg/runtime/...
```

### View per-function coverage

```bash
go tool cover -func=runtime_cov.out
```

### View line-by-line coverage as HTML

```bash
go tool cover -html=runtime_cov.out -o coverage.html
```

### Run a specific test

```bash
go test -v -run TestForDynamic_DistributesAcrossGoroutines ./pkg/runtime/
```

The `-run` flag accepts a regex, so `-run "TestBarrier_.*"` runs every barrier test, for example.
