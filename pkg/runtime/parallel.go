package runtime

import (
	"sync"
	"sync/atomic"
)

// teamContext holds the synchronization state shared by every goroutine that
// participates in a single parallel region.
type teamContext struct {
	barrier *cyclicBarrier
	size    int

	// singleFlag is the election token for Single: the goroutine that wins the
	// CAS from 0 to 1 runs the single body; the others skip it. It is reset to 0
	// after every single so the same region can contain several.
	singleFlag int64

	// workCounter is the shared cursor that worksharing constructs with dynamic
	// distribution (Sections, dynamic For) use to hand out the next unit of
	// work to whichever team goroutine asks. It is reset to 0 at the closing
	// barrier of each construct.
	workCounter int64
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

// newTeam creates a team context sized for the given number of participants,
// with a cyclic barrier so that Barrier() calls from inside the team
// synchronize correctly and can be issued repeatedly within one region.
func newTeam(size int) *teamContext {
	return &teamContext{
		barrier: newCyclicBarrier(size),
		size:    size,
	}
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

// For is a worksharing construct: it is called from inside a parallel region
// by every goroutine of the team, and each caller identified by threadID
// executes the contiguous static chunk of [0, iterations) assigned to it. The
// iteration space is split into team-size blocks of near-equal length, with
// the remainder spread one-per-goroutine across the lowest thread IDs.
// Outside a parallel region (no team) For degrades to running the whole loop
// sequentially in the calling goroutine, with no barrier.
func For(threadID int, body func(int), iterations int) {
	team := getCurrentTeam()
	if team == nil {
		for i := 0; i < iterations; i++ {
			body(i)
		}
		return
	}

	size := team.size
	if iterations > 0 {
		chunk := iterations / size
		rem := iterations % size
		start := threadID*chunk + min(threadID, rem)
		end := start + chunk
		if threadID < rem {
			end++
		}
		for i := start; i < end; i++ {
			body(i)
		}
	}
	Barrier()
}

// ParallelFor is the combined construct: it creates a team and distributes the
// loop across it in a single call. Following the specification, it is exactly
// syntactic sugar for a static For inside a Parallel region.
func ParallelFor(body func(int), iterations int) {
	if iterations <= 0 {
		return
	}
	Parallel(func(threadID int) {
		For(threadID, body, iterations)
	})
}

// ForDynamic is the dynamic-schedule worksharing construct: it is called from
// inside a parallel region by every goroutine of the team, which repeatedly
// claim chunks of chunkSize consecutive iterations from the team's shared
// cursor until the space is exhausted, then synchronize at the implicit
// barrier. Because chunks are handed out on demand, goroutines that finish
// quick work pick up more, balancing uneven iteration costs.
func ForDynamic(body func(int), iterations, chunkSize int) {
	team := getCurrentTeam()
	if team == nil {
		for i := 0; i < iterations; i++ {
			body(i)
		}
		return
	}

	if chunkSize <= 0 {
		chunkSize = 1
	}
	if iterations > 0 {
		for {
			start := atomic.AddInt64(&team.workCounter, int64(chunkSize)) - int64(chunkSize)
			if start >= int64(iterations) {
				break
			}
			end := start + int64(chunkSize)
			if end > int64(iterations) {
				end = int64(iterations)
			}
			for i := start; i < end; i++ {
				body(int(i))
			}
		}
	}
	team.barrier.waitThen(func() {
		atomic.StoreInt64(&team.workCounter, 0)
	})
}

// ParallelForDynamic is the combined construct: it creates a team and runs a
// dynamic-schedule loop across it in a single call. It is exactly a ForDynamic
// worksharing construct inside a Parallel region.
func ParallelForDynamic(body func(int), iterations, chunkSize int) {
	if iterations <= 0 {
		return
	}
	Parallel(func(int) {
		ForDynamic(body, iterations, chunkSize)
	})
}

// Sections is a worksharing construct: it is called from inside a parallel
// region by every goroutine of the team, and the team collectively claims the
// blocks from the shared cursor (dynamic distribution) so each runs exactly
// once on whichever goroutine grabs it. A call ends at the implicit barrier
// that closes the construct.
func Sections(sections []func()) {
	if len(sections) == 0 {
		return
	}

	team := getCurrentTeam()
	if team == nil {
		for _, s := range sections {
			s()
		}
		return
	}

	total := int64(len(sections))
	for {
		idx := atomic.AddInt64(&team.workCounter, 1) - 1
		if idx >= total {
			break
		}
		sections[idx]()
	}
	team.barrier.waitThen(func() {
		atomic.StoreInt64(&team.workCounter, 0)
	})
}

// ParallelSections is the combined construct: it creates a team and distributes
// the section blocks across it in a single call. Following the specification,
// it is exactly a Sections worksharing construct inside a Parallel region.
func ParallelSections(sections []func()) {
	if len(sections) == 0 {
		return
	}
	Parallel(func(int) {
		Sections(sections)
	})
}
