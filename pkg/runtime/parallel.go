package runtime

import (
	"sync"
	"sync/atomic"
)

// teamContext holds the synchronization state shared by every goroutine that
// participates in a single parallel region.
type teamContext struct {
	barrier *sync.WaitGroup
	size    int
}

var (
	// teamMap associates each pool worker (by goroutine ID) with the team
	// context it is currently executing within. The mapping is rewritten every
	// time a worker picks up a new job and cleared when the job finishes.
	teamMap   = make(map[int64]*teamContext)
	teamMapMu sync.RWMutex
)

// registerInTeam binds the calling goroutine to the given team for the
// remainder of its current job. Pool workers invoke this immediately before
// executing a job's body.
func registerInTeam(team *teamContext) {
	gid := getGoroutineID()
	teamMapMu.Lock()
	teamMap[gid] = team
	teamMapMu.Unlock()
}

// unregisterFromTeam clears the calling goroutine's team binding. Invoked by
// the pool worker after a job's body returns, so the same worker can later
// participate in a different team.
func unregisterFromTeam() {
	gid := getGoroutineID()
	teamMapMu.Lock()
	delete(teamMap, gid)
	teamMapMu.Unlock()
}

// getCurrentTeam returns the team context bound to the calling goroutine, or
// nil if the goroutine is not currently executing inside any parallel region.
func getCurrentTeam() *teamContext {
	gid := getGoroutineID()
	teamMapMu.RLock()
	defer teamMapMu.RUnlock()
	return teamMap[gid]
}

// newTeam creates a team context sized for the given number of participants
// and pre-Adds that count to its barrier WaitGroup so that Barrier() calls
// from inside the team synchronize correctly without further setup.
func newTeam(size int) *teamContext {
	t := &teamContext{
		barrier: &sync.WaitGroup{},
		size:    size,
	}
	t.barrier.Add(size)
	return t
}

// Parallel instantiates a team of PoolSize() goroutines, each receiving its
// thread ID, and blocks until every member returns. When
// invoked from inside an already-active parallel region the call is
// serialized: the body executes once with thread ID 0 in a virtual team of
// size 1, mirroring OpenMP's "nested parallelism disabled" default.
func Parallel(body func(int)) {
	// Nested invocation: serialize and run body in the calling goroutine with
	// a transient virtual team of one so that Barrier() and CurrentTeamSize()
	// see consistent values for the nested scope.
	if outer := getCurrentTeam(); outer != nil {
		unregisterFromTeam()
		registerInTeam(newTeam(1))
		body(0)
		unregisterFromTeam()
		registerInTeam(outer)
		return
	}

	p := getPool()
	team := newTeam(p.size)

	var wg sync.WaitGroup
	wg.Add(p.size)

	for tid := 0; tid < p.size; tid++ {
		p.submit(job{
			body:     body,
			threadID: tid,
			team:     team,
			done:     &wg,
		})
	}
	wg.Wait()
}

// For distributes the iteration space across the pool by
// splitting it into PoolSize() contiguous chunks of approximately equal size.
// Each chunk is dispatched as a separate job to a pool worker.
func For(body func(int), iterations int) {
	if iterations <= 0 {
		return
	}

	p := getPool()
	team := newTeam(p.size)

	chunkSize := iterations / p.size
	remainder := iterations % p.size

	var wg sync.WaitGroup
	wg.Add(p.size)

	for tid := 0; tid < p.size; tid++ {
		start := tid * chunkSize
		end := start + chunkSize
		if tid == p.size-1 {
			end += remainder
		}

		chunkStart, chunkEnd := start, end
		p.submit(job{
			body: func(int) {
				for i := chunkStart; i < chunkEnd; i++ {
					body(i)
				}
			},
			threadID: tid,
			team:     team,
			done:     &wg,
		})
	}
	wg.Wait()
}

// ParallelFor is the combined construct that creates a parallel team and
// statically distributes loop iterations in a single call. Semantically
// equivalent to wrapping a static For inside a Parallel region.
func ParallelFor(body func(int), iterations int) {
	if iterations <= 0 {
		return
	}

	p := getPool()
	team := newTeam(p.size)

	chunkSize := iterations / p.size
	remainder := iterations % p.size

	var wg sync.WaitGroup
	wg.Add(p.size)

	for tid := 0; tid < p.size; tid++ {
		start := tid * chunkSize
		end := start + chunkSize
		if tid == p.size-1 {
			end += remainder
		}

		chunkStart, chunkEnd := start, end
		p.submit(job{
			body: func(int) {
				for i := chunkStart; i < chunkEnd; i++ {
					body(i)
				}
			},
			threadID: tid,
			team:     team,
			done:     &wg,
		})
	}
	wg.Wait()
}

// ForDynamic distributes iterations across the pool using a shared atomic
// counter. Each worker repeatedly claims a chunk of "chunkSize" consecutive
// iterations from the counter, executes it, and returns for more until the
// iteration space is exhausted.
func ForDynamic(body func(int), iterations, chunkSize int) {
	if iterations <= 0 {
		return
	}
	if chunkSize <= 0 {
		chunkSize = 1
	}

	p := getPool()
	team := newTeam(p.size)

	var counter int64
	var wg sync.WaitGroup
	wg.Add(p.size)

	for tid := 0; tid < p.size; tid++ {
		p.submit(job{
			body: func(int) {
				for {
					start := atomic.AddInt64(&counter, int64(chunkSize)) - int64(chunkSize)
					if start >= int64(iterations) {
						return
					}
					end := start + int64(chunkSize)
					if end > int64(iterations) {
						end = int64(iterations)
					}
					for i := start; i < end; i++ {
						body(int(i))
					}
				}
			},
			threadID: tid,
			team:     team,
			done:     &wg,
		})
	}
	wg.Wait()
}

// Sections distributes an arbitrary list of independent code blocks across
// the pool. Each block runs exactly once on whichever worker picks it up
// first, using the same atomic-counter dispatch pattern as ForDynamic.
func Sections(sections []func()) {
	if len(sections) == 0 {
		return
	}

	p := getPool()
	workers := p.size
	if workers > len(sections) {
		workers = len(sections)
	}

	team := newTeam(workers)

	var counter int64
	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		p.submit(job{
			body: func(int) {
				for {
					idx := atomic.AddInt64(&counter, 1) - 1
					if idx >= int64(len(sections)) {
						return
					}
					sections[idx]()
				}
			},
			threadID: w,
			team:     team,
			done:     &wg,
		})
	}
	wg.Wait()
}
