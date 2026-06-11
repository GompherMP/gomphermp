package runtime

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestFor_BasicExecution verifies that the team collectively executes every
// iteration exactly once.
func TestFor_BasicExecution(t *testing.T) {
	const iterations = 100
	var counter int64

	Parallel(func(threadID int) {
		For(threadID, func(i int) {
			atomic.AddInt64(&counter, 1)
		}, iterations)
	})

	if counter != iterations {
		t.Errorf("expected %d iterations, got %d", iterations, counter)
	}
}

// TestFor_CorrectIterationValues verifies each index is processed exactly once
// across the team, with the right value.
func TestFor_CorrectIterationValues(t *testing.T) {
	const iterations = 100
	results := make([]int, iterations)

	Parallel(func(threadID int) {
		For(threadID, func(i int) {
			results[i] = i * i
		}, iterations)
	})

	for i := 0; i < iterations; i++ {
		expected := i * i
		if results[i] != expected {
			t.Errorf("results[%d] = %d, expected %d", i, results[i], expected)
		}
	}
}

// TestFor_ZeroIterations verifies an empty iteration space runs no body calls
// but still lets every goroutine reach the loop's implicit barrier.
func TestFor_ZeroIterations(t *testing.T) {
	var counter int64

	Parallel(func(threadID int) {
		For(threadID, func(i int) {
			atomic.AddInt64(&counter, 1)
		}, 0)
	})

	if counter != 0 {
		t.Errorf("expected 0 iterations, got %d", counter)
	}
}

// TestFor_NegativeIterations verifies a negative iteration count is treated as
// empty.
func TestFor_NegativeIterations(t *testing.T) {
	var counter int64

	Parallel(func(threadID int) {
		For(threadID, func(i int) {
			atomic.AddInt64(&counter, 1)
		}, -10)
	})

	if counter != 0 {
		t.Errorf("expected 0 iterations for negative input, got %d", counter)
	}
}

// TestFor_WithClampedPoolSize verifies that For still executes every iteration
// after the pool is clamped to a single worker (team of one).
func TestFor_WithClampedPoolSize(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	const iterations = 10

	for _, requested := range []int{0, -5} {
		SetPoolSize(requested) // clamped to 1
		var counter int64
		Parallel(func(threadID int) {
			For(threadID, func(i int) {
				atomic.AddInt64(&counter, 1)
			}, iterations)
		})
		if counter != iterations {
			t.Errorf("SetPoolSize(%d): expected %d iterations, got %d", requested, iterations, counter)
		}
	}
}

// TestFor_StandaloneRunsSequentially verifies the degraded path: called
// outside any parallel region (no team), For runs the entire loop in the
// calling goroutine, so it remains usable even when misused without an
// enclosing Parallel.
func TestFor_StandaloneRunsSequentially(t *testing.T) {
	const iterations = 50
	results := make([]int, iterations)

	// No enclosing Parallel: getCurrentTeam() is nil.
	For(0, func(i int) {
		results[i] = i + 1
	}, iterations)

	for i := 0; i < iterations; i++ {
		if results[i] != i+1 {
			t.Errorf("results[%d] = %d, expected %d", i, results[i], i+1)
		}
	}
}

// TestFor_FewerIterationsThanThreads verifies correctness when there are fewer
// iterations than team members: every iteration runs exactly once even though
// some goroutines get an empty chunk.
func TestFor_FewerIterationsThanThreads(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(8)

	const iterations = 3
	counts := make([]int64, iterations)

	Parallel(func(threadID int) {
		For(threadID, func(i int) {
			atomic.AddInt64(&counts[i], 1)
		}, iterations)
	})

	for i := 0; i < iterations; i++ {
		if counts[i] != 1 {
			t.Errorf("iteration %d ran %d times, expected exactly 1", i, counts[i])
		}
	}
}

// TestParallel_AllThreadsExecute verifies all threads execute the body.
func TestParallel_AllThreadsExecute(t *testing.T) {
	const threads = 4
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(threads)

	executed := make([]bool, threads)
	var mu sync.Mutex

	Parallel(func(threadID int) {
		mu.Lock()
		executed[threadID] = true
		mu.Unlock()
	})

	for i := 0; i < threads; i++ {
		if !executed[i] {
			t.Errorf("thread %d did not execute", i)
		}
	}
}

// TestParallel_CorrectThreadIDs verifies each thread receives correct ID.
func TestParallel_CorrectThreadIDs(t *testing.T) {
	const threads = 8
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(threads)

	ids := make([]int, threads)
	var mu sync.Mutex

	Parallel(func(threadID int) {
		mu.Lock()
		ids[threadID] = threadID
		mu.Unlock()
	})

	for i := 0; i < threads; i++ {
		if ids[i] != i {
			t.Errorf("ids[%d] = %d, expected %d", i, ids[i], i)
		}
	}
}

// TestParallel_ImplicitBarrier verifies all threads finish before return.
func TestParallel_ImplicitBarrier(t *testing.T) {
	var counter int64

	Parallel(func(threadID int) {
		atomic.AddInt64(&counter, 1)
	})

	// If the implicit barrier works, counter should equal the pool size.
	if counter != int64(PoolSize()) {
		t.Errorf("expected counter=%d, got %d", PoolSize(), counter)
	}
}

// TestParallel_WithClampedPoolSize verifies that Parallel continues to execute
// the body when the pool size is set to invalid values (zero or negative),
// which SetPoolSize internally clamps to the minimum of one.
func TestParallel_WithClampedPoolSize(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	// Pool size 0 -> clamped to 1, body runs exactly once.
	SetPoolSize(0)
	var counter int64
	Parallel(func(threadID int) {
		atomic.AddInt64(&counter, 1)
	})
	if counter != 1 {
		t.Errorf("expected 1 execution after SetPoolSize(0), got %d", counter)
	}

	// Pool size -3 -> clamped to 1 as well.
	SetPoolSize(-3)
	counter = 0
	Parallel(func(threadID int) {
		atomic.AddInt64(&counter, 1)
	})
	if counter != 1 {
		t.Errorf("expected 1 execution after SetPoolSize(-3), got %d", counter)
	}
}

// TestParallel_NestedSerializes verifies that Parallel invoked from inside an
// already-active parallel region runs the body exactly once with thread ID 0
// rather than spawning a second team.
func TestParallel_NestedSerializes(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)
	SetPoolSize(4)

	var (
		innerInvocations int64
		maxInnerID       int64
		mu               sync.Mutex
	)

	Parallel(func(outerID int) {
		Parallel(func(innerID int) {
			atomic.AddInt64(&innerInvocations, 1)
			mu.Lock()
			if int64(innerID) > maxInnerID {
				maxInnerID = int64(innerID)
			}
			mu.Unlock()
		})
	})

	// Each of the 4 outer goroutines runs its nested Parallel exactly once.
	if innerInvocations != 4 {
		t.Errorf("expected 4 nested executions (one per outer thread), got %d", innerInvocations)
	}
	// In serialized nested mode every inner thread ID must be 0.
	if maxInnerID != 0 {
		t.Errorf("expected nested thread ID = 0, observed max = %d", maxInnerID)
	}
}

// TestParallel_NestedTeamSizeIsOne verifies that CurrentTeamSize() observed
// from inside a nested parallel region reports 1, reflecting the virtual
// single-thread team used for serialized nesting.
func TestParallel_NestedTeamSizeIsOne(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)
	SetPoolSize(4)

	var (
		innerSize int64
		mu        sync.Mutex
	)

	Parallel(func(outerID int) {
		Parallel(func(innerID int) {
			mu.Lock()
			innerSize = int64(CurrentTeamSize())
			mu.Unlock()
		})
	})

	if innerSize != 1 {
		t.Errorf("expected CurrentTeamSize() == 1 inside nested Parallel, got %d", innerSize)
	}
}

// TestParallelFor_DistributesWork verifies iterations are distributed correctly.
func TestParallelFor_DistributesWork(t *testing.T) {
	const iterations = 100
	results := make([]bool, iterations)

	ParallelFor(func(i int) {
		results[i] = true
	}, iterations)

	for i := 0; i < iterations; i++ {
		if !results[i] {
			t.Errorf("iteration %d was not executed", i)
		}
	}
}

// TestParallelFor_CorrectValues verifies each iteration receives correct index.
func TestParallelFor_CorrectValues(t *testing.T) {
	const iterations = 50
	results := make([]int, iterations)

	ParallelFor(func(i int) {
		results[i] = i * 2
	}, iterations)

	for i := 0; i < iterations; i++ {
		expected := i * 2
		if results[i] != expected {
			t.Errorf("results[%d] = %d, expected %d", i, results[i], expected)
		}
	}
}

// TestParallelFor_ZeroIterations verifies ParallelFor handles zero iterations.
func TestParallelFor_ZeroIterations(t *testing.T) {
	var counter int64

	ParallelFor(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, 0)

	if counter != 0 {
		t.Errorf("expected 0 iterations, got %d", counter)
	}
}

// TestParallelFor_NegativeIterations verifies ParallelFor rejects negative iteration counts.
// Symmetric counterpart of TestFor_NegativeIterations.
func TestParallelFor_NegativeIterations(t *testing.T) {
	var counter int64

	ParallelFor(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, -7)

	if counter != 0 {
		t.Errorf("expected 0 iterations for negative input, got %d", counter)
	}
}

// TestParallelFor_WithClampedPoolSize verifies that ParallelFor continues to
// execute every iteration after the pool size is clamped to one.
func TestParallelFor_WithClampedPoolSize(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(0)

	const iterations = 10
	var counter int64

	ParallelFor(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations)

	if counter != iterations {
		t.Errorf("expected %d iterations, got %d", iterations, counter)
	}
}

// TestForDynamic_AllIterationsExecute verifies every iteration runs exactly once.
func TestParallelForDynamic_AllIterationsExecute(t *testing.T) {
	const iterations = 100
	counts := make([]int64, iterations)

	ParallelForDynamic(func(i int) {
		atomic.AddInt64(&counts[i], 1)
	}, iterations, 5)

	for i := 0; i < iterations; i++ {
		if counts[i] != 1 {
			t.Errorf("iteration %d executed %d times, expected exactly 1", i, counts[i])
		}
	}
}

// TestForDynamic_CorrectValues verifies each iteration receives its index.
func TestParallelForDynamic_CorrectValues(t *testing.T) {
	const iterations = 50
	results := make([]int, iterations)

	ParallelForDynamic(func(i int) {
		results[i] = i * 2
	}, iterations, 10)

	for i := 0; i < iterations; i++ {
		expected := i * 2
		if results[i] != expected {
			t.Errorf("results[%d] = %d, expected %d", i, results[i], expected)
		}
	}
}

// TestForDynamic_ChunkSizeOne verifies execution with chunkSize=1 (one iteration per claim).
func TestParallelForDynamic_ChunkSizeOne(t *testing.T) {
	const iterations = 30
	var counter int64

	ParallelForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations, 1)

	if counter != iterations {
		t.Errorf("expected %d iterations, got %d", iterations, counter)
	}
}

// TestForDynamic_ChunkSizeLargerThanIterations verifies a single goroutine
// takes all iterations when the chunk size exceeds the iteration count.
func TestParallelForDynamic_ChunkSizeLargerThanIterations(t *testing.T) {
	const iterations = 10
	var counter int64

	ParallelForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations, 1000)

	if counter != iterations {
		t.Errorf("expected %d iterations, got %d", iterations, counter)
	}
}

// TestForDynamic_ZeroIterations verifies the function returns immediately.
func TestParallelForDynamic_ZeroIterations(t *testing.T) {
	var counter int64

	ParallelForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, 0, 5)

	if counter != 0 {
		t.Errorf("expected 0 iterations, got %d", counter)
	}
}

// TestForDynamic_NegativeIterations verifies negative input is rejected.
func TestParallelForDynamic_NegativeIterations(t *testing.T) {
	var counter int64

	ParallelForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, -5, 2)

	if counter != 0 {
		t.Errorf("expected 0 iterations for negative input, got %d", counter)
	}
}

// TestForDynamic_InvalidChunkSize verifies chunkSize <= 0 is corrected to 1.
func TestParallelForDynamic_InvalidChunkSize(t *testing.T) {
	const iterations = 20
	var counter int64

	ParallelForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations, 0)

	if counter != iterations {
		t.Errorf("chunkSize=0: expected %d iterations, got %d", iterations, counter)
	}

	counter = 0
	ParallelForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations, -3)

	if counter != iterations {
		t.Errorf("chunkSize=-3: expected %d iterations, got %d", iterations, counter)
	}
}

// TestForDynamic_WithClampedPoolSize verifies that ForDynamic continues to
// execute every iteration after the pool size is clamped to one by an
// invalid SetPoolSize argument.
func TestParallelForDynamic_WithClampedPoolSize(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(0)

	const iterations = 15
	var counter int64

	ParallelForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations, 4)

	if counter != iterations {
		t.Errorf("expected %d iterations, got %d", iterations, counter)
	}
}

// TestParallelForDynamic_DistributesAcrossGoroutines verifies that work is
// actually dispatched to multiple goroutines, not serialized to one. Each
// iteration sleeps briefly so a single fast goroutine cannot drain the whole
// chunk queue before its teammates start claiming - without that, dynamic
// scheduling is free to (legitimately) run everything on one goroutine, which
// would make the multi-goroutine assertion racy.
func TestParallelForDynamic_DistributesAcrossGoroutines(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(4)

	const iterations = 40
	var (
		seenIDs sync.Map
		counter int64
	)

	ParallelForDynamic(func(i int) {
		seenIDs.Store(getGoroutineID(), struct{}{})
		atomic.AddInt64(&counter, 1)
		time.Sleep(time.Millisecond)
	}, iterations, 1)

	if counter != iterations {
		t.Fatalf("expected %d iterations, got %d", iterations, counter)
	}

	distinct := 0
	seenIDs.Range(func(_, _ any) bool {
		distinct++
		return true
	})
	if distinct < 2 {
		t.Errorf("expected at least 2 distinct goroutines to participate, saw %d", distinct)
	}
}

// TestForDynamic_StressNoRace runs many iterations with the smallest possible
// chunk size to stress the shared atomic counter. Every iteration must run
// exactly once with no losses or duplicates, even under maximum contention.
func TestParallelForDynamic_StressNoRace(t *testing.T) {
	const iterations = 10000
	results := make([]int64, iterations)

	ParallelForDynamic(func(i int) {
		atomic.AddInt64(&results[i], 1)
	}, iterations, 1)

	for i := 0; i < iterations; i++ {
		if results[i] != 1 {
			t.Errorf("iteration %d ran %d times, expected exactly 1", i, results[i])
		}
	}
}

// TestForDynamic_WorksharingInParallel verifies the worksharing path: when
// every goroutine of a parallel team calls ForDynamic, the team collectively
// runs each iteration exactly once by claiming chunks from the shared cursor,
// then synchronizes at the construct's implicit barrier.
func TestForDynamic_WorksharingInParallel(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)
	SetPoolSize(4)

	const iterations = 1000
	counts := make([]int64, iterations)
	var afterLoop int64

	Parallel(func(threadID int) {
		ForDynamic(func(i int) {
			atomic.AddInt64(&counts[i], 1)
		}, iterations, 4)
		atomic.AddInt64(&afterLoop, 1)
	})

	for i := 0; i < iterations; i++ {
		if counts[i] != 1 {
			t.Errorf("iteration %d ran %d times across the team, expected exactly 1", i, counts[i])
		}
	}
	if afterLoop != int64(PoolSize()) {
		t.Errorf("expected all %d goroutines past the loop barrier, got %d", PoolSize(), afterLoop)
	}
}

// TestForDynamic_StandaloneSequential verifies that ForDynamic with no
// surrounding team runs every iteration once, sequentially.
func TestForDynamic_StandaloneSequential(t *testing.T) {
	const iterations = 40
	results := make([]int, iterations)

	ForDynamic(func(i int) {
		results[i] = i + 1
	}, iterations, 5)

	for i := 0; i < iterations; i++ {
		if results[i] != i+1 {
			t.Errorf("results[%d] = %d, expected %d", i, results[i], i+1)
		}
	}
}

// Sections is worksharing: standalone it runs sequentially, while the combined
// ParallelSections provisions a team. The parallel-distribution tests therefore
// exercise ParallelSections; a dedicated test covers the worksharing path of
// Sections called inside an explicit Parallel region.

// TestParallelSections_AllSectionsExecute verifies every section runs exactly once.
func TestParallelSections_AllSectionsExecute(t *testing.T) {
	const total = 6
	counts := make([]int64, total)

	sections := make([]func(), total)
	for i := 0; i < total; i++ {
		i := i
		sections[i] = func() {
			atomic.AddInt64(&counts[i], 1)
		}
	}

	ParallelSections(sections)

	for i := 0; i < total; i++ {
		if counts[i] != 1 {
			t.Errorf("section %d executed %d times, expected exactly 1", i, counts[i])
		}
	}
}

// TestParallelSections_FewerSectionsThanThreads verifies correctness when there
// are fewer sections than team members.
func TestParallelSections_FewerSectionsThanThreads(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(8)
	var counter int64

	ParallelSections([]func(){
		func() { atomic.AddInt64(&counter, 1) },
		func() { atomic.AddInt64(&counter, 1) },
	})

	if counter != 2 {
		t.Errorf("expected 2 executions, got %d", counter)
	}
}

// TestParallelSections_MoreSectionsThanThreads verifies all sections execute
// when there are more sections than team members.
func TestParallelSections_MoreSectionsThanThreads(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(2)
	const total = 20
	var counter int64

	sections := make([]func(), total)
	for i := 0; i < total; i++ {
		sections[i] = func() { atomic.AddInt64(&counter, 1) }
	}

	ParallelSections(sections)

	if counter != int64(total) {
		t.Errorf("expected %d executions, got %d", total, counter)
	}
}

// TestSections_WorksharingInParallel verifies the worksharing path: when every
// goroutine of a parallel team calls Sections, each block still runs exactly
// once across the team (not once per goroutine), and the team synchronizes at
// the construct's implicit barrier.
func TestSections_WorksharingInParallel(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)
	SetPoolSize(4)

	const total = 10
	counts := make([]int64, total)
	var afterSections int64

	sections := make([]func(), total)
	for i := 0; i < total; i++ {
		i := i
		sections[i] = func() { atomic.AddInt64(&counts[i], 1) }
	}

	Parallel(func(threadID int) {
		Sections(sections)
		atomic.AddInt64(&afterSections, 1)
	})

	for i := 0; i < total; i++ {
		if counts[i] != 1 {
			t.Errorf("section %d ran %d times across the team, expected exactly 1", i, counts[i])
		}
	}
	if afterSections != int64(PoolSize()) {
		t.Errorf("expected all %d goroutines past the sections barrier, got %d", PoolSize(), afterSections)
	}
}

// TestSections_StandaloneSequential verifies that Sections called with no
// surrounding team runs every block once, sequentially.
func TestSections_StandaloneSequential(t *testing.T) {
	var counter int64
	Sections([]func(){
		func() { atomic.AddInt64(&counter, 1) },
		func() { atomic.AddInt64(&counter, 1) },
		func() { atomic.AddInt64(&counter, 1) },
	})
	if counter != 3 {
		t.Errorf("expected 3 sequential executions, got %d", counter)
	}
}

// TestSections_EmptyList verifies an empty input returns immediately for both
// the worksharing and combined entry points.
func TestSections_EmptyList(t *testing.T) {
	Sections([]func(){})
	Sections(nil)
	ParallelSections([]func(){})
	ParallelSections(nil)
}

// TestParallelSections_WithClampedPoolSize verifies that every block still runs
// after the pool size is clamped to one.
func TestParallelSections_WithClampedPoolSize(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(0)
	var counter int64

	ParallelSections([]func(){
		func() { atomic.AddInt64(&counter, 1) },
		func() { atomic.AddInt64(&counter, 1) },
		func() { atomic.AddInt64(&counter, 1) },
	})

	if counter != 3 {
		t.Errorf("expected 3 executions, got %d", counter)
	}
}

// TestParallelSections_DifferentBodies verifies each section runs its own function.
func TestParallelSections_DifferentBodies(t *testing.T) {
	var (
		ranA, ranB, ranC bool
		mu               sync.Mutex
	)

	ParallelSections([]func(){
		func() { mu.Lock(); ranA = true; mu.Unlock() },
		func() { mu.Lock(); ranB = true; mu.Unlock() },
		func() { mu.Lock(); ranC = true; mu.Unlock() },
	})

	if !ranA || !ranB || !ranC {
		t.Errorf("expected all sections to run; got ranA=%v ranB=%v ranC=%v", ranA, ranB, ranC)
	}
}
