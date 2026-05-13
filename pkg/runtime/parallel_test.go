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

// TestFor_InvalidNumThreads verifies For handles NumThreads <= 0.
func TestFor_InvalidNumThreads(t *testing.T) {
	// Save and restore original NumThreads
	originalThreads := NumThreads
	defer func() { NumThreads = originalThreads }()

	// Test with NumThreads = 0
	NumThreads = 0

	const iterations = 10
	var counter int64

	For(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations)

	// Verify all iterations were executed
	if counter != iterations {
		t.Errorf("expected %d iterations, got %d", iterations, counter)
	}

	// Verify NumThreads was auto-corrected to 1
	if NumThreads != 1 {
		t.Errorf("expected NumThreads=1 after auto-correct, got %d", NumThreads)
	}

	// Reset and test with negative NumThreads
	NumThreads = -5
	counter = 0

	For(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations)

	if counter != iterations {
		t.Errorf("expected %d iterations with negative NumThreads, got %d", iterations, counter)
	}

	if NumThreads != 1 {
		t.Errorf("expected NumThreads=1 after negative correction, got %d", NumThreads)
	}
}

// TestParallel_AllThreadsExecute verifies all threads execute the body.
func TestParallel_AllThreadsExecute(t *testing.T) {
	const threads = 4
	originalThreads := NumThreads
	defer func() { NumThreads = originalThreads }()

	NumThreads = threads

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
	originalThreads := NumThreads
	defer func() { NumThreads = originalThreads }()

	NumThreads = threads

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

	// If barrier works, counter should equal NumThreads
	if counter != int64(NumThreads) {
		t.Errorf("expected counter=%d, got %d", NumThreads, counter)
	}
}

// TestParallel_InvalidNumThreads verifies Parallel handles NumThreads <= 0.
func TestParallel_InvalidNumThreads(t *testing.T) {
	originalThreads := NumThreads
	defer func() { NumThreads = originalThreads }()

	// Test with NumThreads = 0
	NumThreads = 0

	var executed bool

	Parallel(func(threadID int) {
		executed = true
	})

	if !executed {
		t.Error("Parallel did not execute with NumThreads=0")
	}

	// Verify NumThreads was corrected
	if NumThreads != 1 {
		t.Errorf("expected NumThreads=1, got %d", NumThreads)
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

// TestParallelFor_InvalidNumThreads verifies ParallelFor handles NumThreads <= 0.
func TestParallelFor_InvalidNumThreads(t *testing.T) {
	originalThreads := NumThreads
	defer func() { NumThreads = originalThreads }()

	NumThreads = 0

	const iterations = 10
	var counter int64

	ParallelFor(func(i int) {
		atomic.AddInt64(&counter, 1)
	}, iterations)

	if counter != iterations {
		t.Errorf("expected %d iterations, got %d", iterations, counter)
	}

	if NumThreads != 1 {
		t.Errorf("expected NumThreads=1, got %d", NumThreads)
	}
}
