package runtime

import (
	"sync"
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
		// Anonymous critical section — use global lock
		anonLock.Lock()
		defer anonLock.Unlock()
		body()
	} else {
		// Named critical section — get or create named lock
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

// Single executes body exactly once.
// In a parallel context, the transformer generates coordination code
// (sync.Once + barrier) to ensure single execution across goroutines.
// This runtime function serves as the execution primitive.
// Single executes body exactly once
func Single(body func()) {
	body()
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
func Barrier() {
	team := getCurrentTeam()
	if team == nil {
		// Not in a parallel region, no-op
		return
	}

	team.barrier.Done()
	team.barrier.Wait()
}
