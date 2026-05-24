package runtime

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestTask_BodyExecutes verifies that Task executes its body.
func TestTask_BodyExecutes(t *testing.T) {
	var executed int64

	Taskgroup(func() {
		Task(func() {
			atomic.StoreInt64(&executed, 1)
		})
	})

	if atomic.LoadInt64(&executed) != 1 {
		t.Error("Task body did not execute")
	}
}

// TestTask_IsAsync verifies that Task returns immediately without waiting for body to finish.
func TestTask_IsAsync(t *testing.T) {
	blocker := make(chan struct{})
	returned := make(chan struct{})

	go func() {
		Taskgroup(func() {
			Task(func() { <-blocker })
			close(returned)
		})
	}()

	select {
	case <-returned:
		// Task() returned before body finished — correct
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Task() blocked the caller instead of returning immediately")
	}

	close(blocker)
}

// TestTask_MultipleTasksExecute verifies that multiple tasks all run to completion.
func TestTask_MultipleTasksExecute(t *testing.T) {
	const n = 20
	var counter int64

	Taskgroup(func() {
		for i := 0; i < n; i++ {
			Task(func() {
				atomic.AddInt64(&counter, 1)
			})
		}
	})

	if atomic.LoadInt64(&counter) != n {
		t.Errorf("expected %d tasks to execute, got %d", n, counter)
	}
}

// TestTaskwait_WaitsForDirectChildren verifies Taskwait blocks until all direct
// child tasks have completed.
func TestTaskwait_WaitsForDirectChildren(t *testing.T) {
	var result int64

	Taskgroup(func() {
		Task(func() {
			time.Sleep(20 * time.Millisecond)
			atomic.StoreInt64(&result, 1)
		})

		Taskwait()

		if atomic.LoadInt64(&result) != 1 {
			t.Error("Taskwait returned before child task completed")
		}
	})
}

// TestTaskwait_NoopOutsideTask verifies Taskwait does not block when called
// outside any task context.
func TestTaskwait_NoopOutsideTask(t *testing.T) {
	done := make(chan struct{})

	go func() {
		Taskwait()
		close(done)
	}()

	select {
	case <-done:
		// Returned immediately — correct
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Taskwait blocked outside task context")
	}
}

// TestTaskgroup_WaitsFullSubtree verifies Taskgroup waits for all descendants,
// including tasks spawned by tasks at any nesting depth.
func TestTaskgroup_WaitsFullSubtree(t *testing.T) {
	var deepResult int64

	Taskgroup(func() {
		Task(func() {
			Task(func() {
				Task(func() {
					time.Sleep(20 * time.Millisecond)
					atomic.StoreInt64(&deepResult, 42)
				})
			})
		})
	})

	if atomic.LoadInt64(&deepResult) != 42 {
		t.Error("Taskgroup returned before deeply nested task completed")
	}
}

// TestTaskgroup_GrandchildrenIncluded verifies Taskgroup waits for grandchildren
// while Taskwait would not have.
func TestTaskgroup_GrandchildrenIncluded(t *testing.T) {
	const tasks = 10
	var counter int64

	Taskgroup(func() {
		for i := 0; i < tasks; i++ {
			Task(func() {
				Task(func() {
					atomic.AddInt64(&counter, 1)
				})
			})
		}
	})

	if atomic.LoadInt64(&counter) != tasks {
		t.Errorf("expected %d grandchildren to complete, got counter=%d", tasks, counter)
	}
}

// TestTaskloop_AllIterationsExecute verifies every index in [0, iterations) runs exactly once.
func TestTaskloop_AllIterationsExecute(t *testing.T) {
	const iterations = 50
	results := make([]int64, iterations)

	Taskgroup(func() {
		Taskloop(func(i int) {
			atomic.AddInt64(&results[i], 1)
		}, iterations, 1)
	})

	for i := 0; i < iterations; i++ {
		if results[i] != 1 {
			t.Errorf("iteration %d executed %d times, expected exactly 1", i, results[i])
		}
	}
}

// TestTaskloop_CorrectIterationValues verifies each iteration receives its correct index.
func TestTaskloop_CorrectIterationValues(t *testing.T) {
	const iterations = 30
	results := make([]int, iterations)

	Taskgroup(func() {
		Taskloop(func(i int) {
			results[i] = i * i
		}, iterations, 5)
	})

	for i := 0; i < iterations; i++ {
		if results[i] != i*i {
			t.Errorf("results[%d] = %d, expected %d", i, results[i], i*i)
		}
	}
}

// TestTaskloop_ZeroIterations verifies Taskloop is a no-op for zero iterations.
func TestTaskloop_ZeroIterations(t *testing.T) {
	var counter int64

	Taskgroup(func() {
		Taskloop(func(i int) {
			atomic.AddInt64(&counter, 1)
		}, 0, 1)
	})

	if counter != 0 {
		t.Errorf("expected 0 iterations, got %d", counter)
	}
}

// TestTaskloop_InvalidGrainsize verifies grainsize <= 0 defaults to 1.
func TestTaskloop_InvalidGrainsize(t *testing.T) {
	const iterations = 10

	for _, gs := range []int{0, -1, -5} {
		results := make([]int64, iterations)

		Taskgroup(func() {
			Taskloop(func(i int) {
				atomic.AddInt64(&results[i], 1)
			}, iterations, gs)
		})

		for i := 0; i < iterations; i++ {
			if results[i] != 1 {
				t.Errorf("grainsize=%d: iteration %d ran %d times, expected 1", gs, i, results[i])
			}
		}
	}
}

// TestTaskloop_GrainsizeLargerThanIterations verifies a single task handles all
// iterations when grainsize exceeds the iteration count.
func TestTaskloop_GrainsizeLargerThanIterations(t *testing.T) {
	const iterations = 5
	var counter int64

	Taskgroup(func() {
		Taskloop(func(i int) {
			atomic.AddInt64(&counter, 1)
		}, iterations, 1000)
	})

	if counter != iterations {
		t.Errorf("expected %d iterations, got %d", iterations, counter)
	}
}
