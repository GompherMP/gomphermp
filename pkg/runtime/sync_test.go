package runtime

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestCritical_PreventRaceCondition verifies Critical prevents race conditions.
func TestCritical_PreventRaceCondition(t *testing.T) {
	var counter int64
	const goroutines = 100
	const increments = 1000

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < increments; i++ {
				Critical("", func() {
					// This would race without Critical
					temp := counter
					temp++
					counter = temp
				})
			}
		}()
	}
	wg.Wait()

	expected := int64(goroutines * increments)
	if counter != expected {
		t.Errorf("race condition detected: expected %d, got %d", expected, counter)
	}
}

// TestCritical_NamedLocks verifies named locks are independent.
func TestCritical_NamedLocks(t *testing.T) {
	var counterA, counterB int64
	const iterations = 1000

	var wg sync.WaitGroup

	// Goroutine 1 - uses lock "A"
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			Critical("lockA", func() {
				counterA++
			})
		}
	}()

	// Goroutine 2 - uses lock "B" (should not block on A)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			Critical("lockB", func() {
				counterB++
			})
		}
	}()

	wg.Wait()

	if counterA != iterations {
		t.Errorf("counterA: expected %d, got %d", iterations, counterA)
	}
	if counterB != iterations {
		t.Errorf("counterB: expected %d, got %d", iterations, counterB)
	}
}

// TestCritical_SameNamedLockSerializes verifies same-named locks serialize.
func TestCritical_SameNamedLockSerializes(t *testing.T) {
	var counter int64
	const goroutines = 10
	const increments = 100

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < increments; i++ {
				Critical("sharedLock", func() {
					temp := counter
					temp++
					counter = temp
				})
			}
		}()
	}
	wg.Wait()

	expected := int64(goroutines * increments)
	if counter != expected {
		t.Errorf("expected %d, got %d", expected, counter)
	}
}

// TestCritical_AnonymousVsNamed verifies anonymous and named locks are independent.
func TestCritical_AnonymousVsNamed(t *testing.T) {
	var anonCounter, namedCounter int64
	const iterations = 500

	var wg sync.WaitGroup

	// Goroutine using anonymous lock
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			Critical("", func() {
				anonCounter++
			})
		}
	}()

	// Goroutine using named lock (should run in parallel with anonymous)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			Critical("mylock", func() {
				namedCounter++
			})
		}
	}()

	wg.Wait()

	if anonCounter != iterations {
		t.Errorf("anonCounter: expected %d, got %d", iterations, anonCounter)
	}
	if namedCounter != iterations {
		t.Errorf("namedCounter: expected %d, got %d", iterations, namedCounter)
	}
}

// TestSingle_Executes verifies Single executes the body.
func TestSingle_Executes(t *testing.T) {
	var executed bool

	Single(func() {
		executed = true
	})

	if !executed {
		t.Error("Single did not execute body")
	}
}

// TestSingle_ExecutesMultipleTimes verifies Single can be called multiple times.
func TestSingle_ExecutesMultipleTimes(t *testing.T) {
	var counter int

	Single(func() {
		counter++
	})

	Single(func() {
		counter++
	})

	if counter != 2 {
		t.Errorf("expected counter=2, got %d", counter)
	}
}

// TestMaster_OnlyMasterExecutes verifies only thread 0 executes master block.
func TestMaster_OnlyMasterExecutes(t *testing.T) {
	const goroutines = 4

	var masterCount int64
	var nonMasterCount int64

	var wg sync.WaitGroup
	for id := 0; id < goroutines; id++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			Master(threadID, func() {
				if threadID == 0 {
					atomic.AddInt64(&masterCount, 1)
				} else {
					atomic.AddInt64(&nonMasterCount, 1)
				}
			})
		}(id)
	}
	wg.Wait()

	if masterCount != 1 {
		t.Errorf("expected master to execute once, got %d", masterCount)
	}
	if nonMasterCount != 0 {
		t.Errorf("expected non-master threads to skip, but %d executed", nonMasterCount)
	}
}

// TestMaster_NoImplicitBarrier verifies Master has no barrier (threads don't wait).
func TestMaster_NoImplicitBarrier(t *testing.T) {
	const goroutines = 4

	started := make(chan int, goroutines)
	finished := make(chan int, goroutines)

	var wg sync.WaitGroup
	for id := 0; id < goroutines; id++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			started <- threadID

			Master(threadID, func() {
				// Master thread sleeps
				time.Sleep(100 * time.Millisecond)
			})

			finished <- threadID
		}(id)
	}

	// Wait for all to start
	for i := 0; i < goroutines; i++ {
		<-started
	}

	// Non-master threads should finish quickly (no barrier)
	// Master thread is sleeping, but others shouldn't wait
	finishCount := 0
	timeout := time.After(50 * time.Millisecond)

	for finishCount < goroutines-1 {
		select {
		case id := <-finished:
			if id != 0 {
				finishCount++
			}
		case <-timeout:
			t.Fatal("non-master threads blocked (implicit barrier detected)")
		}
	}

	wg.Wait()
}

// TestMaster_CorrectThreadIDCheck verifies Master checks threadID correctly.
func TestMaster_CorrectThreadIDCheck(t *testing.T) {
	executed := make(map[int]bool)
	var mu sync.Mutex

	var wg sync.WaitGroup
	for id := 0; id < 8; id++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			Master(threadID, func() {
				mu.Lock()
				executed[threadID] = true
				mu.Unlock()
			})
		}(id)
	}
	wg.Wait()

	// Only thread 0 should have executed
	if len(executed) != 1 {
		t.Errorf("expected 1 thread to execute, got %d: %v", len(executed), executed)
	}
	if !executed[0] {
		t.Error("thread 0 did not execute")
	}
}

// TestMaster_AllThreadsContinue verifies all threads continue after Master.
func TestMaster_AllThreadsContinue(t *testing.T) {
	const goroutines = 4

	var afterMaster int64

	var wg sync.WaitGroup
	for id := 0; id < goroutines; id++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			Master(threadID, func() {
				// Only master executes this
			})

			// But ALL threads execute this
			atomic.AddInt64(&afterMaster, 1)
		}(id)
	}
	wg.Wait()

	if afterMaster != goroutines {
		t.Errorf("expected all %d threads to continue, only %d did", goroutines, afterMaster)
	}
}

// TestBarrier_SynchronizesGoroutines verifies all goroutines wait at barrier.
func TestBarrier_SynchronizesGoroutines(t *testing.T) {
	originalThreads := NumThreads
	defer func() { NumThreads = originalThreads }()
	NumThreads = 4

	var phase1Complete int64
	var phase2Start int64
	var raceDetected int64

	Parallel(func(threadID int) {
		// Phase 1 - all goroutines do this
		atomic.AddInt64(&phase1Complete, 1)

		// Barrier - all must reach here before any continue
		Barrier()

		// Phase 2 - should only start after ALL finished phase 1
		if atomic.LoadInt64(&phase1Complete) != int64(NumThreads) {
			atomic.StoreInt64(&raceDetected, 1)
		}
		atomic.AddInt64(&phase2Start, 1)
	})

	if atomic.LoadInt64(&raceDetected) != 0 {
		t.Error("phase 2 started before all goroutines finished phase 1")
	}
	if phase1Complete != int64(NumThreads) {
		t.Errorf("expected %d in phase 1, got %d", NumThreads, phase1Complete)
	}
	if phase2Start != int64(NumThreads) {
		t.Errorf("expected %d in phase 2, got %d", NumThreads, phase2Start)
	}
}

// TestBarrier_OutsideParallel verifies Barrier is no-op outside parallel region.
func TestBarrier_OutsideParallel(t *testing.T) {
	// Should not panic or block
	Barrier()
}

// TestBarrier_DifferentTeamSizes verifies barrier works with various team sizes.
func TestBarrier_DifferentTeamSizes(t *testing.T) {
	originalThreads := NumThreads
	defer func() { NumThreads = originalThreads }()

	sizes := []int{1, 2, 4, 8}

	for _, size := range sizes {
		NumThreads = size
		var counter int64

		Parallel(func(threadID int) {
			atomic.AddInt64(&counter, 1)
			Barrier()
		})

		if counter != int64(size) {
			t.Errorf("team size %d: expected counter=%d, got %d", size, size, counter)
		}
	}
}

// TestBarrier_NoDeadlock verifies barrier completes in reasonable time.
func TestBarrier_NoDeadlock(t *testing.T) {
	originalThreads := NumThreads
	defer func() { NumThreads = originalThreads }()
	NumThreads = 8

	done := make(chan bool, 1)

	go func() {
		Parallel(func(threadID int) {
			Barrier()
		})
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("barrier deadlocked")
	}
}

// TestBarrier_OrderOfOperations verifies work after barrier sees work before barrier.
func TestBarrier_OrderOfOperations(t *testing.T) {
	originalThreads := NumThreads
	defer func() { NumThreads = originalThreads }()
	NumThreads = 4

	values := make([]int, NumThreads)
	var sumAfter int64

	Parallel(func(threadID int) {
		// Each goroutine writes to its own slot
		values[threadID] = threadID * 10

		Barrier()

		// After barrier, all values should be visible
		// Sum should be 0 + 10 + 20 + 30 = 60
		total := 0
		for _, v := range values {
			total += v
		}
		atomic.AddInt64(&sumAfter, int64(total))
	})

	// Each goroutine sees the total of 60
	// Total sumAfter = 60 * NumThreads = 240
	expected := int64(60 * NumThreads)
	if sumAfter != expected {
		t.Errorf("expected sumAfter=%d, got %d", expected, sumAfter)
	}
}
