package runtime

import (
	"sync"
	"testing"
)

// TestAtomicAddInt_SingleThread verifies the basic arithmetic and return
// value: each call adds delta and returns the new value, matching the
// semantics of `*p += delta` expressed as an expression.
func TestAtomicAddInt_SingleThread(t *testing.T) {
	x := 10

	if got := AtomicAddInt(&x, 5); got != 15 {
		t.Errorf("AtomicAddInt(+5) returned %d, want 15", got)
	}
	if x != 15 {
		t.Errorf("x = %d after add, want 15", x)
	}
	if got := AtomicAddInt(&x, -7); got != 8 {
		t.Errorf("AtomicAddInt(-7) returned %d, want 8", got)
	}
	if x != 8 {
		t.Errorf("x = %d after subtract, want 8", x)
	}
}

// TestAtomicAddInt_NoLostUpdates is the core correctness test: many goroutines
// incrementing a shared int concurrently must produce exactly the expected
// total. A non-atomic `x++` would lose updates under the race detector and the
// final count would fall short.
func TestAtomicAddInt_NoLostUpdates(t *testing.T) {
	const goroutines = 100
	const incrementsEach = 1000

	var x int
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < incrementsEach; i++ {
				AtomicAddInt(&x, 1)
			}
		}()
	}
	wg.Wait()

	want := goroutines * incrementsEach
	if x != want {
		t.Errorf("final counter = %d, want %d (lost updates)", x, want)
	}
}

// TestAtomicStoreLoadInt verifies that a stored value is read back exactly,
// covering the write and read directive forms together.
func TestAtomicStoreLoadInt(t *testing.T) {
	var x int

	AtomicStoreInt(&x, 42)
	if got := AtomicLoadInt(&x); got != 42 {
		t.Errorf("AtomicLoadInt after store(42) = %d, want 42", got)
	}

	AtomicStoreInt(&x, -1)
	if got := AtomicLoadInt(&x); got != -1 {
		t.Errorf("AtomicLoadInt after store(-1) = %d, want -1", got)
	}
}

// TestAtomicStoreLoadInt_Concurrent verifies that concurrent stores and loads
// never observe a torn value: every load must return one of the values that
// was actually stored, never a partial mix. Run under -race this also asserts
// the absence of a data race on the shared int.
func TestAtomicStoreLoadInt_Concurrent(t *testing.T) {
	var x int
	AtomicStoreInt(&x, 1)

	var wg sync.WaitGroup

	// Writers alternate between two distinct values.
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				AtomicStoreInt(&x, base)
			}
		}(w%2 + 1) // stores 1 or 2
	}

	// Readers must only ever see a fully-written value (1 or 2).
	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				if v := AtomicLoadInt(&x); v != 1 && v != 2 {
					t.Errorf("observed torn/unexpected value %d", v)
					return
				}
			}
		}()
	}

	wg.Wait()
}
