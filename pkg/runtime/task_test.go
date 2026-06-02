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

// TestTaskwait_MultipleChildren verifies Taskwait blocks until all direct
// child tasks have completed, not just the first one.
func TestTaskwait_MultipleChildren(t *testing.T) {
	const n = 5
	results := make([]int64, n)

	Taskgroup(func() {
		for i := 0; i < n; i++ {
			i := i
			Task(func() {
				time.Sleep(10 * time.Millisecond)
				atomic.StoreInt64(&results[i], 1)
			})
		}

		Taskwait()

		for i := 0; i < n; i++ {
			if atomic.LoadInt64(&results[i]) != 1 {
				t.Errorf("child %d not done after Taskwait", i)
			}
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

// TestTaskloop_NegativeIterations verifies Taskloop is a no-op for negative iterations.
func TestTaskloop_NegativeIterations(t *testing.T) {
	var counter int64

	Taskgroup(func() {
		Taskloop(func(i int) {
			atomic.AddInt64(&counter, 1)
		}, -10, 1)
	})

	if counter != 0 {
		t.Errorf("expected 0 iterations for negative input, got %d", counter)
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

// TestTaskloop_SingleIteration verifies that a single-iteration Taskloop executes
// exactly once and delivers index 0 to the body.
func TestTaskloop_SingleIteration(t *testing.T) {
	var counter int64
	var receivedIndex int64

	Taskgroup(func() {
		Taskloop(func(i int) {
			atomic.AddInt64(&counter, 1)
			atomic.StoreInt64(&receivedIndex, int64(i))
		}, 1, 1)
	})

	if counter != 1 {
		t.Errorf("expected exactly 1 iteration, got %d", counter)
	}
	if receivedIndex != 0 {
		t.Errorf("expected index 0, got %d", receivedIndex)
	}
}

// TestTask_OutsideTaskContext verifies Task can be called outside any Taskgroup.
// The spawned goroutine must still execute even with no parent task context.
func TestTask_OutsideTaskContext(t *testing.T) {
	done := make(chan struct{})

	Task(func() {
		close(done)
	})

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Task body did not execute when called outside a task context")
	}
}

// TestTaskwait_DoesNotWaitGrandchildren verifies that Taskwait returns as soon
// as direct children are done, even if their children (grandchildren) are still
// running. Designed as a deadlock trap: if Taskwait mistakenly waited for the
// grandchild, unblock would never be closed and the test would hang.
func TestTaskwait_DoesNotWaitGrandchildren(t *testing.T) {
	grandchildSubmitted := make(chan struct{})
	unblock := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)
		Taskgroup(func() {
			Task(func() {
				Task(func() { <-unblock }) // grandchild blocks indefinitely
				close(grandchildSubmitted) // signal: grandchild is live
			})
			<-grandchildSubmitted // ensure grandchild submitted before Taskwait
			Taskwait()            // must return once the direct child is done
			close(unblock)        // only reachable if Taskwait did not wait for grandchild
		})
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("deadlock: Taskwait likely waited for grandchildren instead of direct children only")
	}
}

// TestTaskgroup_Nested verifies that a Taskgroup invoked inside a Task waits
// for its own subtasks before returning, so the outer task can observe the
// inner results immediately after the inner Taskgroup call.
func TestTaskgroup_Nested(t *testing.T) {
	var innerResult, outerResult int64

	Taskgroup(func() {
		Task(func() {
			Taskgroup(func() {
				Task(func() {
					time.Sleep(10 * time.Millisecond)
					atomic.StoreInt64(&innerResult, 1)
				})
			})
			// Inner Taskgroup has returned, so innerResult must be 1.
			atomic.StoreInt64(&outerResult, atomic.LoadInt64(&innerResult))
		})
	})

	if atomic.LoadInt64(&innerResult) != 1 {
		t.Error("inner task did not complete before inner Taskgroup returned")
	}
	if atomic.LoadInt64(&outerResult) != 1 {
		t.Error("outer task did not observe inner result after inner Taskgroup")
	}
}

// TestTaskgroup_NoDeadlock verifies Taskgroup completes in finite time when
// spawning a large number of tasks.
func TestTaskgroup_NoDeadlock(t *testing.T) {
	done := make(chan struct{})

	go func() {
		defer close(done)
		Taskgroup(func() {
			for i := 0; i < 100; i++ {
				Task(func() {})
			}
		})
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Taskgroup deadlocked with 100 concurrent tasks")
	}
}

// TestTask_StressNoRace spawns 1000 tasks, each incrementing a dedicated slot,
// and verifies every slot is incremented exactly once with no data races.
func TestTask_StressNoRace(t *testing.T) {
	const n = 1000
	results := make([]int64, n)

	Taskgroup(func() {
		for i := 0; i < n; i++ {
			i := i
			Task(func() {
				atomic.AddInt64(&results[i], 1)
			})
		}
	})

	for i := 0; i < n; i++ {
		if results[i] != 1 {
			t.Errorf("task %d ran %d times, expected exactly 1", i, results[i])
		}
	}
}
