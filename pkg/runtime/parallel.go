package runtime

import (
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// NumThreads is the size of the goroutine team for parallel regions.
var NumThreads = runtime.NumCPU()

// teamContext holds synchronization state for a parallel region.
type teamContext struct {
	barrier *sync.WaitGroup
	size    int
}

var (
	// teamMap associates goroutine IDs with their team context
	teamMap   = make(map[int64]*teamContext)
	teamMapMu sync.RWMutex
)

// getGoroutineID returns the current goroutine's ID by parsing the stack.
// This is a well-known Go idiom used by production libraries.
func getGoroutineID() int64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.ParseInt(idField, 10, 64)
	return id
}

// registerInTeam associates this goroutine with the given team context.
func registerInTeam(team *teamContext) {
	gid := getGoroutineID()
	teamMapMu.Lock()
	teamMap[gid] = team
	teamMapMu.Unlock()
}

// unregisterFromTeam removes this goroutine's team association.
func unregisterFromTeam() {
	gid := getGoroutineID()
	teamMapMu.Lock()
	delete(teamMap, gid)
	teamMapMu.Unlock()
}

// getCurrentTeam returns the team context for the current goroutine, or nil.
func getCurrentTeam() *teamContext {
	gid := getGoroutineID()
	teamMapMu.RLock()
	defer teamMapMu.RUnlock()
	return teamMap[gid]
}

// For distributes loop iterations across goroutines.
// It divides the iteration space [0, iterations) into NumThreads chunks
// and assigns one chunk to each goroutine.
func For(body func(int), iterations int) {
	// Handle edge cases
	if iterations <= 0 {
		return
	}

	if NumThreads <= 0 {
		NumThreads = 1
	}

	chunkSize := iterations / NumThreads
	remainder := iterations % NumThreads

	var wg sync.WaitGroup

	for threadID := 0; threadID < NumThreads; threadID++ {
		wg.Add(1)

		start := threadID * chunkSize
		end := start + chunkSize

		if threadID == NumThreads-1 {
			end += remainder
		}

		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				body(i)
			}
		}(start, end)
	}

	wg.Wait()
}

// Parallel creates a team of goroutines, each executing body concurrently.
// Each goroutine receives its thread ID.
// Sets up team context so Barrier() and other team-aware functions work.
// Waits for all goroutines to complete before returning (implicit barrier).
func Parallel(body func(int)) {
	if NumThreads <= 0 {
		NumThreads = 1
	}

	// Create team context with barrier sized for the team
	team := &teamContext{
		barrier: &sync.WaitGroup{},
		size:    NumThreads,
	}
	team.barrier.Add(NumThreads)

	var wg sync.WaitGroup
	for threadID := 0; threadID < NumThreads; threadID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Register this goroutine with the team
			registerInTeam(team)
			defer unregisterFromTeam()

			body(id)
		}(threadID)
	}

	wg.Wait()
}

// ParallelFor combines Parallel and For — creates a team and distributes iterations.
// This is a convenience function equivalent to calling Parallel with For inside.
func ParallelFor(body func(int), iterations int) {
	if iterations <= 0 {
		return
	}

	if NumThreads <= 0 {
		NumThreads = 1
	}

	chunkSize := iterations / NumThreads
	remainder := iterations % NumThreads

	var wg sync.WaitGroup

	for threadID := 0; threadID < NumThreads; threadID++ {
		wg.Add(1)

		start := threadID * chunkSize
		end := start + chunkSize

		if threadID == NumThreads-1 {
			end += remainder
		}

		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				body(i)
			}
		}(start, end)
	}

	wg.Wait()
}

// ForDynamic distributes iterations across goroutines using a shared work queue.
// Each goroutine repeatedly claims a chunk of `chunkSize` consecutive iterations
// from the queue, executes it, and returns for more until the iteration space
// is exhausted.
func ForDynamic(body func(int), iterations, chunkSize int) {
	if iterations <= 0 {
		return
	}
	if chunkSize <= 0 {
		chunkSize = 1
	}
	if NumThreads <= 0 {
		NumThreads = 1
	}

	var counter int64
	var wg sync.WaitGroup

	for threadID := 0; threadID < NumThreads; threadID++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
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
		}()
	}

	wg.Wait()
}

// Sections distributes independent code blocks across goroutines. Each block
// is executed by exactly one goroutine, with assignment made dynamically as
// goroutines become available.
func Sections(sections []func()) {
	if len(sections) == 0 {
		return
	}
	if NumThreads <= 0 {
		NumThreads = 1
	}

	workers := NumThreads
	if workers > len(sections) {
		workers = len(sections)
	}

	var counter int64
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				idx := atomic.AddInt64(&counter, 1) - 1
				if idx >= int64(len(sections)) {
					return
				}
				sections[idx]()
			}
		}()
	}

	wg.Wait()
}
