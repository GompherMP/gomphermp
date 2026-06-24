package runtime

import (
	"runtime"
	"sync"
)

// job represents a single unit of work submitted to the worker pool.
// The worker that picks the job registers itself in "team" before executing
// "body", so team-aware primitives like Barrier() can resolve the right
// synchronization context.
type job struct {
	body     func(int)
	threadID int
	team     *teamContext
	done     *sync.WaitGroup
}

// pool is the persistent goroutine pool that backs every work-distribution
// primitive in the runtime. A fixed number of goroutines (one per worker)
// stay alive for the lifetime of the pool, consuming jobs from a shared
// channel. Reusing these workers across parallel regions avoids the
// creation/destruction overhead of spawning fresh goroutines on every call.
type pool struct {
	jobs chan job
	size int
}

var (
	// poolMu guards swaps of the "current" pointer when SetPoolSize is called.
	poolMu sync.RWMutex

	// current is the active pool.
	current *pool
)

// init creates the worker pool at package load time with a size matching
// runtime.GOMAXPROCS(0). This guarantees the pool is ready before any caller
// (including the first invocation from main) requests parallel work.
func init() {
	current = newPool(runtime.GOMAXPROCS(0))
}

// newPool allocates a pool with the requested number of workers and launches
// them. Sizes less than or equal to zero are clamped to one to keep the
// runtime usable in degraded configurations.
func newPool(size int) *pool {
	if size <= 0 {
		size = 1
	}
	p := &pool{
		jobs: make(chan job, size*4),
		size: size,
	}
	for w := 0; w < size; w++ {
		go poolWorker(p)
	}
	return p
}

// poolWorker is the long-lived goroutine that consumes jobs from the pool's
// channel. It exits cleanly when the channel is closed (which happens during
// SetPoolSize) so resized pools do not leak goroutines.
func poolWorker(p *pool) {
	for j := range p.jobs {
		if j.team != nil {
			registerInTeam(j.team)
		}
		j.body(j.threadID)
		if j.team != nil {
			unregisterFromTeam()
		}
		j.done.Done()
	}
}

// submit enqueues a job for execution by one of the pool workers. Blocks if
// the channel buffer is full (which is bounded by 4 * pool size).
func (p *pool) submit(j job) {
	p.jobs <- j
}

// getPool returns the active pool atomically with respect to SetPoolSize.
func getPool() *pool {
	poolMu.RLock()
	defer poolMu.RUnlock()
	return current
}

// PoolSize returns the size of the active worker pool. By default this matches
// runtime.GOMAXPROCS(0), but it can be reconfigured via SetPoolSize.
func PoolSize() int {
	return getPool().size
}

// CurrentTeamSize returns the size of the team to which the calling goroutine
// belongs.
func CurrentTeamSize() int {
	team := getCurrentTeam()
	if team == nil {
		return 1
	}
	return team.size
}

// SetPoolSize replaces the active pool with a new one of the requested size.
// In production, prefer setting GOMAXPROCS before the program starts to control
// the initial pool size.
func SetPoolSize(n int) {
	poolMu.Lock()
	defer poolMu.Unlock()
	if current != nil {
		close(current.jobs)
	}
	current = newPool(n)
}
