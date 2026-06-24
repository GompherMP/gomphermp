package runtime

import (
	"sync"
	"sync/atomic"
)

var (
	// anonLock protects anonymous critical sections
	anonLock sync.Mutex

	// namedLocks maps lock names to their mutexes
	namedLocks   = make(map[string]*sync.Mutex)
	namedLocksMu sync.Mutex
)

// Critical provides mutual exclusion for the given block.
func Critical(name string, body func()) {
	if name == "" {
		// Anonymous critical section - use global lock
		anonLock.Lock()
		defer anonLock.Unlock()
		body()
	} else {
		// Named critical section - get or create named lock
		lock := getNamedLock(name)
		lock.Lock()
		defer lock.Unlock()
		body()
	}
}

// getNamedLock retrieves or creates a mutex for the given name.
func getNamedLock(name string) *sync.Mutex {
	namedLocksMu.Lock()
	defer namedLocksMu.Unlock()

	if lock, exists := namedLocks[name]; exists {
		return lock
	}

	lock := &sync.Mutex{}
	namedLocks[name] = lock
	return lock
}

// Single executes body on exactly one goroutine of the current team while the
// others skip it, then synchronizes the whole team at the implicit barrier
// that closes the construct (OpenMP single semantics).
//
// Election is a single atomic CAS on the team's singleFlag: the first goroutine
// to flip it from 0 to 1 runs the body. The implicit barrier that closes the
// construct also resets the token (the last arriver clears it under the barrier
// lock), so a region may contain any number of single blocks back to back.
// Outside a parallel region (no team) Single simply runs the body.
func Single(body func()) {
	team := getCurrentTeam()
	if team == nil {
		body()
		return
	}

	if atomic.CompareAndSwapInt64(&team.singleFlag, 0, 1) {
		body()
	}
	// The implicit barrier closes the construct; the last goroutine to arrive
	// clears the token under the barrier lock, so the reset is visible to all
	// before any of them can reach the next single.
	team.barrier.waitThen(func() {
		atomic.StoreInt64(&team.singleFlag, 0)
	})
}

// Master executes body only on the master thread (threadID 0).
// Unlike Single, there is no implicit barrier
func Master(threadID int, body func()) {
	if threadID == 0 {
		body()
	}
}

// Barrier waits for all goroutines in the current team to reach this point.
// All goroutines in the team must call Barrier before any can proceed.
//
// The barrier is cyclic: it resets after every full rendezvous, so a single
// parallel region may contain many barriers - the explicit //gompher barrier
// as well as the implicit barriers that close each worksharing construct (for,
// sections, single). Outside a parallel region the call is a no-op.
func Barrier() {
	team := getCurrentTeam()
	if team == nil {
		// Not in a parallel region, no-op
		return
	}
	team.barrier.wait()
}

// cyclicBarrier is a reusable barrier for a fixed number of participants. Once
// all size goroutines have called wait, they are all released and the barrier
// re-arms for the next round. This is the difference from a one-shot
// sync.WaitGroup: the same barrier can be used repeatedly within one parallel
// region, which the OpenMP worksharing model requires.
type cyclicBarrier struct {
	mu    sync.Mutex
	cond  *sync.Cond
	size  int
	count int
	gen   uint64
}

func newCyclicBarrier(size int) *cyclicBarrier {
	b := &cyclicBarrier{size: size}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// wait blocks until size goroutines have arrived.
func (b *cyclicBarrier) wait() {
	b.waitThen(nil)
}

// waitThen blocks until size goroutines have arrived, then releases them all.
// If beforeRelease is non-nil, the last goroutine to arrive runs it while
// still holding the barrier's lock, before any waiter is woken. This gives
// worksharing constructs a safe place to reset their shared team state (the
// section/iteration counter, the single token): the reset is published to
// every goroutine before any of them can reach the next construct.
//
// The generation counter lets a goroutine released in round N ignore spurious
// wakeups and avoids the "lost wakeup" race: a waiter only proceeds once gen
// has advanced past the value it observed on entry.
func (b *cyclicBarrier) waitThen(beforeRelease func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	gen := b.gen
	b.count++
	if b.count == b.size {
		// Last to arrive: run the reset (if any), then open the gate and
		// re-arm for the next round.
		if beforeRelease != nil {
			beforeRelease()
		}
		b.count = 0
		b.gen++
		b.cond.Broadcast()
		return
	}
	for gen == b.gen {
		b.cond.Wait()
	}
}
