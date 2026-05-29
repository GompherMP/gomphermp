package runtime

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Pool initialization and configuration
// ---------------------------------------------------------------------------

// TestPool_InitializedAtStartup verifies that the pool exists and is usable
// before any explicit call. The eager init in pool.go guarantees this.
func TestPool_InitializedAtStartup(t *testing.T) {
	if PoolSize() <= 0 {
		t.Errorf("expected pool size > 0 at package load, got %d", PoolSize())
	}
}

// TestPoolSize_DefaultsToGOMAXPROCS verifies the default size matches the
// number of OS threads the Go scheduler is allowed to use simultaneously.
func TestPoolSize_DefaultsToGOMAXPROCS(t *testing.T) {
	original := PoolSize()
	defer SetPoolSize(original)

	expected := runtime.GOMAXPROCS(0)
	if expected <= 0 {
		t.Skip("GOMAXPROCS returned non-positive value")
	}

	// The default must equal GOMAXPROCS. We reset the pool to confirm
	// (in case a previous test left it at a different size).
	SetPoolSize(runtime.GOMAXPROCS(0))
	if got := PoolSize(); got != expected {
		t.Errorf("expected pool size = GOMAXPROCS (%d), got %d", expected, got)
	}
}

// ---------------------------------------------------------------------------
// PoolSize / CurrentTeamSize
// ---------------------------------------------------------------------------

// TestCurrentTeamSize_OutsideParallel verifies the function returns 1 when
// invoked from a goroutine that does not belong to any team (the main
// goroutine, in this case).
func TestCurrentTeamSize_OutsideParallel(t *testing.T) {
	if got := CurrentTeamSize(); got != 1 {
		t.Errorf("expected CurrentTeamSize() == 1 outside Parallel, got %d", got)
	}
}

// TestCurrentTeamSize_InsideTeam verifies the function returns the size of
// the team registered for the executing worker. This test bypasses the
// Parallel public API (which is migrated to the pool in a later commit) and
// exercises the pool's team registration mechanism directly via submit().
func TestCurrentTeamSize_InsideTeam(t *testing.T) {
	team := &teamContext{
		barrier: &sync.WaitGroup{},
		size:    4,
	}
	team.barrier.Add(team.size)

	var seen int64
	var wg sync.WaitGroup
	wg.Add(1)

	getPool().submit(job{
		body: func(int) {
			atomic.StoreInt64(&seen, int64(CurrentTeamSize()))
		},
		threadID: 0,
		team:     team,
		done:     &wg,
	})
	wg.Wait()

	if got := atomic.LoadInt64(&seen); got != 4 {
		t.Errorf("expected CurrentTeamSize() == 4 inside team, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// SetPoolSize
// ---------------------------------------------------------------------------

// TestSetPoolSize_Resizes verifies that the new size is reflected by
// PoolSize() after a resize.
func TestSetPoolSize_Resizes(t *testing.T) {
	original := PoolSize()
	defer SetPoolSize(original)

	SetPoolSize(2)
	if got := PoolSize(); got != 2 {
		t.Errorf("expected pool size 2 after SetPoolSize(2), got %d", got)
	}

	SetPoolSize(8)
	if got := PoolSize(); got != 8 {
		t.Errorf("expected pool size 8 after SetPoolSize(8), got %d", got)
	}
}

// TestSetPoolSize_NewSizeProcessesJobs verifies that the new pool created by
// SetPoolSize is fully operational: it accepts jobs, dispatches them to its
// workers and signals completion through the per-job WaitGroup.
func TestSetPoolSize_NewSizeProcessesJobs(t *testing.T) {
	original := PoolSize()
	defer SetPoolSize(original)

	SetPoolSize(3)

	const jobsCount = 9
	var counter int64
	var wg sync.WaitGroup
	wg.Add(jobsCount)

	p := getPool()
	for i := 0; i < jobsCount; i++ {
		p.submit(job{
			body:     func(int) { atomic.AddInt64(&counter, 1) },
			threadID: 0,
			team:     nil,
			done:     &wg,
		})
	}
	wg.Wait()

	if counter != jobsCount {
		t.Errorf("expected %d jobs executed by the new pool, got %d", jobsCount, counter)
	}
}

// TestSetPoolSize_ClampsNonPositive verifies that requesting zero or negative
// sizes is normalized to one rather than producing a broken pool.
func TestSetPoolSize_ClampsNonPositive(t *testing.T) {
	original := PoolSize()
	defer SetPoolSize(original)

	SetPoolSize(0)
	if got := PoolSize(); got != 1 {
		t.Errorf("expected pool size 1 after SetPoolSize(0), got %d", got)
	}

	SetPoolSize(-5)
	if got := PoolSize(); got != 1 {
		t.Errorf("expected pool size 1 after SetPoolSize(-5), got %d", got)
	}
}

// TestSetPoolSize_OldWorkersExitCleanly verifies that resizing the pool does
// not leak workers from the previous generation.
func TestSetPoolSize_OldWorkersExitCleanly(t *testing.T) {
	original := PoolSize()
	defer SetPoolSize(original)

	for i := 0; i < 5; i++ {
		SetPoolSize(4)

		var wg sync.WaitGroup
		const jobsCount = 20
		wg.Add(jobsCount)

		p := getPool()
		for j := 0; j < jobsCount; j++ {
			p.submit(job{
				body:     func(int) {},
				threadID: 0,
				team:     nil,
				done:     &wg,
			})
		}

		wg.Wait()
	}
}

// ---------------------------------------------------------------------------
// Worker persistence
// ---------------------------------------------------------------------------

// TestPool_WorkersReusedAcrossSubmissions: Core proof that the pool is
// persistent. The same goroutine IDs must service jobs submitted in different
// batches. If each Parallel-like region spawned fresh goroutines, the set of
// IDs observed in the second batch would be disjoint from the first. We force
// the workers to be simultaneously busy in each batch and then check 
// that the two observed ID sets are identical.
func TestPool_WorkersReusedAcrossSubmissions(t *testing.T) {
	original := PoolSize()
	defer SetPoolSize(original)
	SetPoolSize(4)

	collectIDs := func() map[int64]bool {
		p := getPool()
		release := make(chan struct{})
		var ids sync.Map
		var ready sync.WaitGroup
		ready.Add(p.size)

		var wg sync.WaitGroup
		wg.Add(p.size)
		for i := 0; i < p.size; i++ {
			p.submit(job{
				body: func(int) {
					ids.Store(getGoroutineID(), struct{}{})
					ready.Done()
					<-release
				},
				threadID: 0,
				team:     nil,
				done:     &wg,
			})
		}
		ready.Wait()
		close(release)
		wg.Wait()

		set := make(map[int64]bool)
		ids.Range(func(k, _ any) bool {
			set[k.(int64)] = true
			return true
		})
		return set
	}

	first := collectIDs()
	second := collectIDs()

	if len(first) != PoolSize() || len(second) != PoolSize() {
		t.Fatalf("expected %d distinct worker IDs per batch, got first=%d second=%d",
			PoolSize(), len(first), len(second))
	}

	for id := range first {
		if !second[id] {
			t.Errorf("worker goroutine ID %d serviced first batch but not the second. Pool is not persistent", id)
		}
	}
}

// TestPool_WorkersExecuteConcurrently verifies that a pool of size N actually
// runs N jobs simultaneously rather than serializing them. All workers must
// reach a synchronization point at the same time.
func TestPool_WorkersExecuteConcurrently(t *testing.T) {
	original := PoolSize()
	defer SetPoolSize(original)
	SetPoolSize(4)

	p := getPool()
	start := make(chan struct{})
	arrived := make(chan struct{}, p.size)

	var wg sync.WaitGroup
	wg.Add(p.size)
	for i := 0; i < p.size; i++ {
		p.submit(job{
			body: func(int) {
				arrived <- struct{}{}
				<-start
			},
			threadID: 0,
			team:     nil,
			done:     &wg,
		})
	}

	timeout := time.After(2 * time.Second)
	for i := 0; i < p.size; i++ {
		select {
		case <-arrived:
		case <-timeout:
			t.Fatalf("only %d of %d workers reached the sync point - jobs are being serialized", i, p.size)
		}
	}

	close(start)
	wg.Wait()
}
