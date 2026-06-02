package runtime

import (
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

// TestTaskWithDepend_OutBeforeIn verifies that a depend(in:x) task waits for
// the depend(out:x) task that produces x to complete first.
func TestTaskWithDepend_OutBeforeIn(t *testing.T) {
	var x int64
	var result int64
	xAddr := uintptr(unsafe.Pointer(&x))

	Taskgroup(func() {
		TaskWithDepend(func() {
			time.Sleep(20 * time.Millisecond)
			atomic.StoreInt64(&x, 42)
		}, nil, []uintptr{xAddr}, nil) // out:x

		TaskWithDepend(func() {
			atomic.StoreInt64(&result, atomic.LoadInt64(&x))
		}, []uintptr{xAddr}, nil, nil) // in:x
	})

	if result != 42 {
		t.Errorf("in task read x=%d before out task wrote it; expected 42", result)
	}
}

// TestTaskWithDepend_InoutChain verifies that inout dependencies serialize a
// pipeline of tasks on the same token in submission order.
func TestTaskWithDepend_InoutChain(t *testing.T) {
	var buf int64
	bufAddr := uintptr(unsafe.Pointer(&buf))

	Taskgroup(func() {
		TaskWithDepend(func() {
			atomic.StoreInt64(&buf, 1)
		}, nil, []uintptr{bufAddr}, nil) // out

		TaskWithDepend(func() {
			atomic.AddInt64(&buf, 1)
		}, nil, nil, []uintptr{bufAddr}) // inout

		TaskWithDepend(func() {
			atomic.AddInt64(&buf, 1)
		}, nil, nil, []uintptr{bufAddr}) // inout
	})

	if buf != 3 {
		t.Errorf("expected buf=3 after pipeline, got %d", buf)
	}
}

// TestTaskWithDepend_ConsecutiveWriters verifies that a second depend(out:x) task
// waits for the first depend(out:x) task on the same address to complete.
func TestTaskWithDepend_ConsecutiveWriters(t *testing.T) {
	var x int64
	xAddr := uintptr(unsafe.Pointer(&x))

	Taskgroup(func() {
		TaskWithDepend(func() {
			time.Sleep(20 * time.Millisecond)
			atomic.StoreInt64(&x, 1)
		}, nil, []uintptr{xAddr}, nil) // first out:x

		TaskWithDepend(func() {
			atomic.StoreInt64(&x, atomic.LoadInt64(&x)+1)
		}, nil, []uintptr{xAddr}, nil) // second out:x — must wait for first
	})

	if x != 2 {
		t.Errorf("expected x=2 after two serialized writers, got %d", x)
	}
}

// TestTaskWithDepend_IndependentTokens verifies tasks on different dependency
// tokens are not serialized and both execute correctly.
func TestTaskWithDepend_IndependentTokens(t *testing.T) {
	var x, y int64
	xAddr := uintptr(unsafe.Pointer(&x))
	yAddr := uintptr(unsafe.Pointer(&y))

	Taskgroup(func() {
		TaskWithDepend(func() {
			atomic.StoreInt64(&x, 1)
		}, nil, []uintptr{xAddr}, nil)

		TaskWithDepend(func() {
			atomic.StoreInt64(&y, 2)
		}, nil, []uintptr{yAddr}, nil)
	})

	if x != 1 {
		t.Errorf("expected x=1, got %d", x)
	}
	if y != 2 {
		t.Errorf("expected y=2, got %d", y)
	}
}

// TestTaskWithDepend_InWithNoWriter verifies that a depend(in:x) task proceeds
// immediately when no prior depend(out:x) task has claimed that address.
func TestTaskWithDepend_InWithNoWriter(t *testing.T) {
	var x int64 = 42
	var result int64
	xAddr := uintptr(unsafe.Pointer(&x))

	Taskgroup(func() {
		TaskWithDepend(func() {
			atomic.StoreInt64(&result, atomic.LoadInt64(&x))
		}, []uintptr{xAddr}, nil, nil) // in:x with no prior out:x
	})

	if result != 42 {
		t.Errorf("expected result=42, got %d", result)
	}
}

// TestTaskWithDepend_MultipleReaders verifies that two depend(in:x) tasks are
// both unblocked once the shared depend(out:x) predecessor completes, and both
// observe the written value.
func TestTaskWithDepend_MultipleReaders(t *testing.T) {
	var x int64
	var r1, r2 int64
	xAddr := uintptr(unsafe.Pointer(&x))

	Taskgroup(func() {
		TaskWithDepend(func() {
			time.Sleep(20 * time.Millisecond)
			atomic.StoreInt64(&x, 42)
		}, nil, []uintptr{xAddr}, nil) // out:x

		TaskWithDepend(func() {
			atomic.StoreInt64(&r1, atomic.LoadInt64(&x))
		}, []uintptr{xAddr}, nil, nil) // in:x

		TaskWithDepend(func() {
			atomic.StoreInt64(&r2, atomic.LoadInt64(&x))
		}, []uintptr{xAddr}, nil, nil) // in:x
	})

	if r1 != 42 {
		t.Errorf("reader 1 got x=%d, expected 42", r1)
	}
	if r2 != 42 {
		t.Errorf("reader 2 got x=%d, expected 42", r2)
	}
}

// TestTaskWithDepend_ReadersDoNotBlockEachOther verifies that two depend(in:x)
// tasks run concurrently once their shared out:x writer finishes. Implemented
// as a deadlock trap: each reader waits for the other to signal; if the runtime
// serialized readers they would deadlock, causing the test to time out.
func TestTaskWithDepend_ReadersDoNotBlockEachOther(t *testing.T) {
	var x int64
	xAddr := uintptr(unsafe.Pointer(&x))

	r1Active := make(chan struct{})
	r2Active := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)
		Taskgroup(func() {
			TaskWithDepend(func() {
				atomic.StoreInt64(&x, 1)
			}, nil, []uintptr{xAddr}, nil) // out:x

			// r1 signals it started, then waits for r2 — would deadlock if serialized.
			TaskWithDepend(func() {
				close(r1Active)
				<-r2Active
			}, []uintptr{xAddr}, nil, nil) // in:x

			// r2 signals it started, then waits for r1 — would deadlock if serialized.
			TaskWithDepend(func() {
				close(r2Active)
				<-r1Active
			}, []uintptr{xAddr}, nil, nil) // in:x
		})
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("deadlock: in-tasks appear to be serialized with each other")
	}
}

// TestTaskWithDepend_InoutAfterMultipleReaders verifies that an inout:x task
// waits for all concurrent in:x predecessors before executing.
func TestTaskWithDepend_InoutAfterMultipleReaders(t *testing.T) {
	var x int64
	xAddr := uintptr(unsafe.Pointer(&x))
	var r1Done, r2Done, inoutSawBoth int64

	Taskgroup(func() {
		TaskWithDepend(func() {
			atomic.StoreInt64(&x, 1)
		}, nil, []uintptr{xAddr}, nil) // out:x

		TaskWithDepend(func() {
			time.Sleep(10 * time.Millisecond)
			atomic.StoreInt64(&r1Done, 1)
		}, []uintptr{xAddr}, nil, nil) // in:x — reader 1

		TaskWithDepend(func() {
			time.Sleep(20 * time.Millisecond)
			atomic.StoreInt64(&r2Done, 1)
		}, []uintptr{xAddr}, nil, nil) // in:x — reader 2

		TaskWithDepend(func() {
			if atomic.LoadInt64(&r1Done) == 1 && atomic.LoadInt64(&r2Done) == 1 {
				atomic.StoreInt64(&inoutSawBoth, 1)
			}
			atomic.AddInt64(&x, 1)
		}, nil, nil, []uintptr{xAddr}) // inout:x — must wait for both readers
	})

	if atomic.LoadInt64(&inoutSawBoth) != 1 {
		t.Error("inout task ran before both in-readers completed")
	}
	if atomic.LoadInt64(&x) != 2 {
		t.Errorf("expected x=2 after out+inout, got %d", x)
	}
}

// TestTaskWithDepend_LongInoutChain verifies that a sequence of 10 inout:x tasks
// serializes in submission order, producing a final value equal to 10.
func TestTaskWithDepend_LongInoutChain(t *testing.T) {
	var buf int64
	bufAddr := uintptr(unsafe.Pointer(&buf))

	Taskgroup(func() {
		for i := int64(1); i <= 10; i++ {
			i := i
			TaskWithDepend(func() {
				atomic.StoreInt64(&buf, i)
			}, nil, nil, []uintptr{bufAddr})
		}
	})

	if atomic.LoadInt64(&buf) != 10 {
		t.Errorf("expected buf=10 after serialized chain, got %d", buf)
	}
}

// TestTaskWithDepend_MultipleInTokens verifies a task with in:x and in:y waits
// for both writers before reading the values they produce.
func TestTaskWithDepend_MultipleInTokens(t *testing.T) {
	var x, y, result int64
	xAddr := uintptr(unsafe.Pointer(&x))
	yAddr := uintptr(unsafe.Pointer(&y))

	Taskgroup(func() {
		TaskWithDepend(func() {
			time.Sleep(20 * time.Millisecond)
			atomic.StoreInt64(&x, 10)
		}, nil, []uintptr{xAddr}, nil) // out:x

		TaskWithDepend(func() {
			time.Sleep(30 * time.Millisecond)
			atomic.StoreInt64(&y, 20)
		}, nil, []uintptr{yAddr}, nil) // out:y

		TaskWithDepend(func() {
			atomic.StoreInt64(&result, atomic.LoadInt64(&x)+atomic.LoadInt64(&y))
		}, []uintptr{xAddr, yAddr}, nil, nil) // in:x, in:y
	})

	if result != 30 {
		t.Errorf("expected result=30 (x+y after both writers), got %d", result)
	}
}

// TestTaskWithDepend_MultipleOutTokens verifies that a task declaring out:x and
// out:y blocks subsequent in:x and in:y readers on both addresses.
func TestTaskWithDepend_MultipleOutTokens(t *testing.T) {
	var x, y, rx, ry int64
	xAddr := uintptr(unsafe.Pointer(&x))
	yAddr := uintptr(unsafe.Pointer(&y))

	Taskgroup(func() {
		TaskWithDepend(func() {
			time.Sleep(20 * time.Millisecond)
			atomic.StoreInt64(&x, 5)
			atomic.StoreInt64(&y, 7)
		}, nil, []uintptr{xAddr, yAddr}, nil) // out:x, out:y

		TaskWithDepend(func() {
			atomic.StoreInt64(&rx, atomic.LoadInt64(&x))
		}, []uintptr{xAddr}, nil, nil) // in:x

		TaskWithDepend(func() {
			atomic.StoreInt64(&ry, atomic.LoadInt64(&y))
		}, []uintptr{yAddr}, nil, nil) // in:y
	})

	if rx != 5 {
		t.Errorf("rx: expected 5, got %d", rx)
	}
	if ry != 7 {
		t.Errorf("ry: expected 7, got %d", ry)
	}
}

// TestTaskWithDepend_OutThenInThenOut verifies the three-stage pipeline
// out → in → out: the second writer must observe that the reader completed
// before it runs, which in turn observed the first writer.
func TestTaskWithDepend_OutThenInThenOut(t *testing.T) {
	var x int64
	xAddr := uintptr(unsafe.Pointer(&x))
	var phase1, phase2, phase3 int64

	Taskgroup(func() {
		TaskWithDepend(func() {
			time.Sleep(10 * time.Millisecond)
			atomic.StoreInt64(&x, 1)
			atomic.StoreInt64(&phase1, 1)
		}, nil, []uintptr{xAddr}, nil) // first out:x

		TaskWithDepend(func() {
			if atomic.LoadInt64(&phase1) == 1 {
				atomic.StoreInt64(&phase2, atomic.LoadInt64(&x))
			}
		}, []uintptr{xAddr}, nil, nil) // in:x

		TaskWithDepend(func() {
			atomic.StoreInt64(&phase3, atomic.LoadInt64(&phase2))
			atomic.StoreInt64(&x, 2)
		}, nil, []uintptr{xAddr}, nil) // second out:x
	})

	if atomic.LoadInt64(&x) != 2 {
		t.Errorf("expected x=2 after second writer, got %d", x)
	}
	if atomic.LoadInt64(&phase2) != 1 {
		t.Errorf("reader did not observe first writer (phase2=%d)", phase2)
	}
	if atomic.LoadInt64(&phase3) != 1 {
		t.Errorf("second writer did not observe reader (phase3=%d)", phase3)
	}
}

// TestTaskWithDepend_InAfterInout verifies that a depend(in:x) task submitted
// after a depend(inout:x) task waits for the inout task to complete, because
// inout registers itself as the new writer.
func TestTaskWithDepend_InAfterInout(t *testing.T) {
	var x, result int64
	xAddr := uintptr(unsafe.Pointer(&x))

	Taskgroup(func() {
		TaskWithDepend(func() {
			time.Sleep(20 * time.Millisecond)
			atomic.StoreInt64(&x, 42)
		}, nil, nil, []uintptr{xAddr}) // inout:x — acts as writer

		TaskWithDepend(func() {
			atomic.StoreInt64(&result, atomic.LoadInt64(&x))
		}, []uintptr{xAddr}, nil, nil) // in:x — must see x=42
	})

	if result != 42 {
		t.Errorf("expected result=42, got %d", result)
	}
}

// TestTaskWithDepend_NoDeadlock verifies that a chain of inout-dependent tasks
// completes in finite time with no deadlock.
func TestTaskWithDepend_NoDeadlock(t *testing.T) {
	done := make(chan struct{})

	go func() {
		defer close(done)
		var x int64
		xAddr := uintptr(unsafe.Pointer(&x))

		Taskgroup(func() {
			for i := 0; i < 20; i++ {
				TaskWithDepend(func() {
					atomic.AddInt64(&x, 1)
				}, nil, nil, []uintptr{xAddr})
			}
		})
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("TaskWithDepend chain deadlocked")
	}
}

// TestTaskWithDepend_StressNoRace runs 50 inout tasks on the same token using
// a non-atomic increment, relying entirely on dependency serialization for
// correctness. Any missed dependency would produce a wrong final count.
func TestTaskWithDepend_StressNoRace(t *testing.T) {
	const n = 50
	var counter int64
	counterAddr := uintptr(unsafe.Pointer(&counter))

	Taskgroup(func() {
		for i := 0; i < n; i++ {
			TaskWithDepend(func() {
				// Non-atomic: correct only if inout serializes all tasks.
				v := counter
				v++
				counter = v
			}, nil, nil, []uintptr{counterAddr})
		}
	})

	if counter != n {
		t.Errorf("expected counter=%d, got %d (data race in dependency ordering)", n, counter)
	}
}
