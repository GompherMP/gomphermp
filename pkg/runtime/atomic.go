package runtime

import (
	"sync/atomic"
	"unsafe"
)

// This functions back the //gompher atomic directive. Go's sync/atomic package
// exposes lock-free operations only for fixed-width integer types (int32,
// int64, ...), but it has no entry point for the platform-dependent int type.
// The helpers below bridge that gap so the transformer can emit atomic
// operations over ordinary int variables, preserving the idiomatic Go style of
// writing `var counter int`.
func AtomicAddInt(p *int, delta int) int {
	return int(atomic.AddInt64((*int64)(unsafe.Pointer(p)), int64(delta)))
}

// AtomicLoadInt atomically reads *p. It backs the read side of
// `//gompher atomic read` (v = x), guaranteeing the load is not torn by a
// concurrent write.
func AtomicLoadInt(p *int) int {
	return int(atomic.LoadInt64((*int64)(unsafe.Pointer(p))))
}

// AtomicStoreInt atomically writes v to *p. It backs `//gompher atomic write`
// (x = e), guaranteeing the store is observed as a single indivisible update.
func AtomicStoreInt(p *int, v int) {
	atomic.StoreInt64((*int64)(unsafe.Pointer(p)), int64(v))
}
