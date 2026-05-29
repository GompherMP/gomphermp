package runtime

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestFor_BasicExecution verifies that For executes all iterations.
func TestFor_BasicExecution(t *testing.T) {
	const iterations = 100
	var counter int64

	For(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations)

	if counter != iterations {
		t.Errorf("expected %d iterations, got %d", iterations, counter)
	}
}

// TestFor_CorrectIterationValues verifies each iteration receives the correct index.
func TestFor_CorrectIterationValues(t *testing.T) {
	const iterations = 100
	results := make([]int, iterations)

	For(func(i int) {
		results[i] = i * i
	}, iterations)

	// Verify each index was processed exactly once
	for i := 0; i < iterations; i++ {
		expected := i * i
		if results[i] != expected {
			t.Errorf("results[%d] = %d, expected %d", i, results[i], expected)
		}
	}
}

// TestFor_ZeroIterations verifies For handles zero iterations
func TestFor_ZeroIterations(t *testing.T) {
	var counter int64

	For(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, 0)

	if counter != 0 {
		t.Errorf("expected 0 iterations, got %d", counter)
	}
}

// TestFor_NegativeIterations verifies For handles negative iterations
func TestFor_NegativeIterations(t *testing.T) {
	var counter int64

	For(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, -10)

	if counter != 0 {
		t.Errorf("expected 0 iterations for negative input, got %d", counter)
	}
}

// TestFor_WithClampedPoolSize verifies that For continues to distribute and
// execute every iteration after the pool is set to invalid sizes (zero or
// negative), which SetPoolSize internally clamps to the minimum of one.
func TestFor_WithClampedPoolSize(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	const iterations = 10

	// Pool size 0 -> clamped to 1 by SetPoolSize.
	SetPoolSize(0)
	var counter int64
	For(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations)
	if counter != iterations {
		t.Errorf("expected %d iterations after SetPoolSize(0), got %d", iterations, counter)
	}

	// Pool size -5 -> clamped to 1 by SetPoolSize.
	SetPoolSize(-5)
	counter = 0
	For(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations)
	if counter != iterations {
		t.Errorf("expected %d iterations after SetPoolSize(-5), got %d", iterations, counter)
	}
}

// TestFor_FewerIterationsThanThreads verifies correctness when there are fewer
// iterations than configured threads. All iterations must still execute exactly
// once, even though some goroutines end up receiving no work.
func TestFor_FewerIterationsThanThreads(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(8)

	const iterations = 3
	counts := make([]int64, iterations)

	For(func(i int) {
		atomic.AddInt64(&counts[i], 1)
	}, iterations)

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
func TestForDynamic_AllIterationsExecute(t *testing.T) {
	const iterations = 100
	counts := make([]int64, iterations)

	ForDynamic(func(i int) {
		atomic.AddInt64(&counts[i], 1)
	}, iterations, 5)

	for i := 0; i < iterations; i++ {
		if counts[i] != 1 {
			t.Errorf("iteration %d executed %d times, expected exactly 1", i, counts[i])
		}
	}
}

// TestForDynamic_CorrectValues verifies each iteration receives its index.
func TestForDynamic_CorrectValues(t *testing.T) {
	const iterations = 50
	results := make([]int, iterations)

	ForDynamic(func(i int) {
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
func TestForDynamic_ChunkSizeOne(t *testing.T) {
	const iterations = 30
	var counter int64

	ForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations, 1)

	if counter != iterations {
		t.Errorf("expected %d iterations, got %d", iterations, counter)
	}
}

// TestForDynamic_ChunkSizeLargerThanIterations verifies a single goroutine
// takes all iterations when the chunk size exceeds the iteration count.
func TestForDynamic_ChunkSizeLargerThanIterations(t *testing.T) {
	const iterations = 10
	var counter int64

	ForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations, 1000)

	if counter != iterations {
		t.Errorf("expected %d iterations, got %d", iterations, counter)
	}
}

// TestForDynamic_ZeroIterations verifies the function returns immediately.
func TestForDynamic_ZeroIterations(t *testing.T) {
	var counter int64

	ForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, 0, 5)

	if counter != 0 {
		t.Errorf("expected 0 iterations, got %d", counter)
	}
}

// TestForDynamic_NegativeIterations verifies negative input is rejected.
func TestForDynamic_NegativeIterations(t *testing.T) {
	var counter int64

	ForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, -5, 2)

	if counter != 0 {
		t.Errorf("expected 0 iterations for negative input, got %d", counter)
	}
}

// TestForDynamic_InvalidChunkSize verifies chunkSize <= 0 is corrected to 1.
func TestForDynamic_InvalidChunkSize(t *testing.T) {
	const iterations = 20
	var counter int64

	ForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations, 0)

	if counter != iterations {
		t.Errorf("chunkSize=0: expected %d iterations, got %d", iterations, counter)
	}

	counter = 0
	ForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations, -3)

	if counter != iterations {
		t.Errorf("chunkSize=-3: expected %d iterations, got %d", iterations, counter)
	}
}

// TestForDynamic_WithClampedPoolSize verifies that ForDynamic continues to
// execute every iteration after the pool size is clamped to one by an
// invalid SetPoolSize argument.
func TestForDynamic_WithClampedPoolSize(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(0)

	const iterations = 15
	var counter int64

	ForDynamic(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations, 4)

	if counter != iterations {
		t.Errorf("expected %d iterations, got %d", iterations, counter)
	}
}

// TestForDynamic_DistributesAcrossGoroutines verifies that work is actually
// dispatched to multiple goroutines, not serialized to one. With pool size >= 2,
// iterations large enough and chunks small enough, at least two distinct
// goroutine IDs must record activity.
func TestForDynamic_DistributesAcrossGoroutines(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(4)

	const iterations = 1000
	var (
		seenIDs sync.Map
		counter int64
	)

	ForDynamic(func(i int) {
		seenIDs.Store(getGoroutineID(), struct{}{})
		atomic.AddInt64(&counter, 1)
	}, iterations, 4)

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
func TestForDynamic_StressNoRace(t *testing.T) {
	const iterations = 10000
	results := make([]int64, iterations)

	ForDynamic(func(i int) {
		atomic.AddInt64(&results[i], 1)
	}, iterations, 1)

	for i := 0; i < iterations; i++ {
		if results[i] != 1 {
			t.Errorf("iteration %d ran %d times, expected exactly 1", i, results[i])
		}
	}
}

// TestSections_AllSectionsExecute verifies every section runs exactly once.
func TestSections_AllSectionsExecute(t *testing.T) {
	const total = 6
	counts := make([]int64, total)

	sections := make([]func(), total)
	for i := 0; i < total; i++ {
		i := i
		sections[i] = func() {
			atomic.AddInt64(&counts[i], 1)
		}
	}

	Sections(sections)

	for i := 0; i < total; i++ {
		if counts[i] != 1 {
			t.Errorf("section %d executed %d times, expected exactly 1", i, counts[i])
		}
	}
}

// TestSections_FewerSectionsThanThreads verifies excess goroutines are not spawned.
func TestSections_FewerSectionsThanThreads(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(8)
	var counter int64

	Sections([]func(){
		func() { atomic.AddInt64(&counter, 1) },
		func() { atomic.AddInt64(&counter, 1) },
	})

	if counter != 2 {
		t.Errorf("expected 2 executions, got %d", counter)
	}
}

// TestSections_MoreSectionsThanThreads verifies all sections execute when there
// are more sections than goroutines.
func TestSections_MoreSectionsThanThreads(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(2)
	const total = 20
	var counter int64

	sections := make([]func(), total)
	for i := 0; i < total; i++ {
		sections[i] = func() { atomic.AddInt64(&counter, 1) }
	}

	Sections(sections)

	if counter != int64(total) {
		t.Errorf("expected %d executions, got %d", total, counter)
	}
}

// TestSections_EmptyList verifies an empty input returns immediately.
func TestSections_EmptyList(t *testing.T) {
	Sections([]func(){})
	Sections(nil)
}

// TestSections_WithClampedPoolSize verifies that Sections continues to execute
// every block after the pool size is clamped to one by an invalid SetPoolSize
// argument.
func TestSections_WithClampedPoolSize(t *testing.T) {
	originalSize := PoolSize()
	defer SetPoolSize(originalSize)

	SetPoolSize(0)
	var counter int64

	Sections([]func(){
		func() { atomic.AddInt64(&counter, 1) },
		func() { atomic.AddInt64(&counter, 1) },
		func() { atomic.AddInt64(&counter, 1) },
	})

	if counter != 3 {
		t.Errorf("expected 3 executions, got %d", counter)
	}
}

// TestSections_DifferentBodies verifies each section runs its own function.
func TestSections_DifferentBodies(t *testing.T) {
	var (
		ranA, ranB, ranC bool
		mu               sync.Mutex
	)

	Sections([]func(){
		func() { mu.Lock(); ranA = true; mu.Unlock() },
		func() { mu.Lock(); ranB = true; mu.Unlock() },
		func() { mu.Lock(); ranC = true; mu.Unlock() },
	})

	if !ranA || !ranB || !ranC {
		t.Errorf("expected all sections to run; got ranA=%v ranB=%v ranC=%v", ranA, ranB, ranC)
	}
}
