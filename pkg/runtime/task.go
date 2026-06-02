package runtime

import (
	"sync"
	"sync/atomic"
)

// taskgroupDepth tracks nested Taskgroup call depth so resetDeps is only
// called when the outermost group exits (all tasks are guaranteed finished).
var taskgroupDepth int64

// taskHandle holds synchronization state for a single task.
type taskHandle struct {
	done     chan struct{} // closed when body finishes
	mu       sync.Mutex
	children []*taskHandle
}

var (
	// taskMap associates goroutine IDs with their current task handle.
	taskMap   = make(map[int64]*taskHandle)
	taskMapMu sync.RWMutex
)

// newHandle allocates and initialises a taskHandle.
func newHandle() *taskHandle {
	return &taskHandle{
		done: make(chan struct{}),
	}
}

// currentTask returns the taskHandle for the current goroutine, or nil.
func currentTask() *taskHandle {
	gid := getGoroutineID()
	taskMapMu.RLock()
	defer taskMapMu.RUnlock()
	return taskMap[gid]
}

// registerTask associates this goroutine with the given task handle.
func registerTask(h *taskHandle) {
	gid := getGoroutineID()
	taskMapMu.Lock()
	taskMap[gid] = h
	taskMapMu.Unlock()
}

// unregisterTask removes this goroutine's task association.
func unregisterTask() {
	gid := getGoroutineID()
	taskMapMu.Lock()
	delete(taskMap, gid)
	taskMapMu.Unlock()
}

// waitSubtree recursively waits for h and all its descendants to finish.
// Safe to call after h.done is already closed — proceeds immediately to children.
func waitSubtree(h *taskHandle) {
	<-h.done
	h.mu.Lock()
	children := make([]*taskHandle, len(h.children))
	copy(children, h.children)
	h.mu.Unlock()
	for _, child := range children {
		waitSubtree(child)
	}
}

// Task submits body as an asynchronous task.
// It is attached to the current task as a child, if one exists.
// Task returns immediately; body runs concurrently.
func Task(body func()) {
	h := newHandle()

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
		body()
	}()
}

// Taskwait blocks until all direct children of the current task have finished.
// Grandchildren and deeper descendants are not waited on.
// No-op if called outside a task context.
func Taskwait() {
	h := currentTask()
	if h == nil {
		return
	}

	h.mu.Lock()
	children := make([]*taskHandle, len(h.children))
	copy(children, h.children)
	h.mu.Unlock()

	for _, child := range children {
		<-child.done
	}
}

// Taskgroup executes body and blocks until all tasks spawned within it —
// including tasks spawned by those tasks at any depth — have finished.
func Taskgroup(body func()) {
	atomic.AddInt64(&taskgroupDepth, 1)
	h := newHandle()
	registerTask(h)
	body()
	unregisterTask()
	close(h.done)
	waitSubtree(h)
	if atomic.AddInt64(&taskgroupDepth, -1) == 0 {
		resetDeps()
	}
}

// Taskloop distributes [0, iterations) as tasks, one per chunk of grainsize.
// grainsize defaults to 1 if <= 0.
func Taskloop(body func(int), iterations, grainsize int) {
	if iterations <= 0 {
		return
	}
	if grainsize <= 0 {
		grainsize = 1
	}

	for start := 0; start < iterations; start += grainsize {
		end := start + grainsize
		if end > iterations {
			end = iterations
		}
		s, e := start, end
		Task(func() {
			for i := s; i < e; i++ {
				body(i)
			}
		})
	}
}
