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
