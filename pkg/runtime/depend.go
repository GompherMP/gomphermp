package runtime

import (
	"sync"
)

// depEntry tracks the completion signals for a single dependency token (variable address).
type depEntry struct {
	writerDone  chan struct{}   // last out/inout task's done channel, nil if none
	readersDone []chan struct{} // active in tasks' done channels
}

var (
	depRegistry   = make(map[uintptr]*depEntry)
	depRegistryMu sync.Mutex
)

// getOrCreateEntry retrieves or creates a depEntry for the given address token.
func getOrCreateEntry(addr uintptr) *depEntry {
	if e, ok := depRegistry[addr]; ok {
		return e
	}
	e := &depEntry{}
	depRegistry[addr] = e
	return e
}

// isClosed reports whether ch has been closed, without blocking.
func isClosed(ch chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

// claimDeps atomically collects the signals this task must wait for and
// registers done in the registry as writer or reader for each address token.
// Waiting on the returned signals happens after unlock, inside the goroutine.
// Stale signals from already-finished tasks are pruned during this call.
func claimDeps(done chan struct{}, ins, outs, inouts []uintptr) []chan struct{} {
	depRegistryMu.Lock()
	defer depRegistryMu.Unlock()

	var signals []chan struct{}

	// in: wait for current writer only; register as reader so future writers wait for us.
	for _, addr := range ins {
		e := getOrCreateEntry(addr)
		if e.writerDone != nil {
			if isClosed(e.writerDone) {
				e.writerDone = nil // prune: writer already finished
			} else {
				signals = append(signals, e.writerDone)
			}
		}
		e.readersDone = append(e.readersDone, done)
	}

	// out: wait for current writer and all active readers; replace as new writer.
	for _, addr := range outs {
		e := getOrCreateEntry(addr)
		if e.writerDone != nil && !isClosed(e.writerDone) {
			signals = append(signals, e.writerDone)
		}
		for _, ch := range e.readersDone {
			if !isClosed(ch) {
				signals = append(signals, ch)
			}
		}
		e.writerDone = done
		e.readersDone = nil
	}

	// inout: same as out - serialises with all prior readers and writers.
	for _, addr := range inouts {
		e := getOrCreateEntry(addr)
		if e.writerDone != nil && !isClosed(e.writerDone) {
			signals = append(signals, e.writerDone)
		}
		for _, ch := range e.readersDone {
			if !isClosed(ch) {
				signals = append(signals, ch)
			}
		}
		e.writerDone = done
		e.readersDone = nil
	}

	return signals
}

// resetDeps clears the dependency registry. Safe to call only once all tasks
// that registered dependency tokens have finished (guaranteed at Taskgroup exit).
func resetDeps() {
	depRegistryMu.Lock()
	depRegistry = make(map[uintptr]*depEntry)
	depRegistryMu.Unlock()
}

// TaskWithDepend submits body as a task with data-flow dependency ordering.
// ins, outs, inouts are variable addresses used as dependency tokens - they are
// never dereferenced; the address is a correlation key only.
// The task is submitted immediately; dependency waiting happens inside the goroutine.
func TaskWithDepend(body func(), ins, outs, inouts []uintptr) {
	h := newHandle()

	signals := claimDeps(h.done, ins, outs, inouts)

	parent := currentTask()
	if parent != nil {
		parent.mu.Lock()
		parent.children = append(parent.children, h)
		parent.mu.Unlock()
	}

	go func() {
		registerTask(h)
		defer close(h.done)
		defer unregisterTask()

		for _, sig := range signals {
			<-sig
		}

		body()
	}()
}
