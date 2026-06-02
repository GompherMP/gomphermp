package runtime

import (
	"runtime"
	"strconv"
	"strings"
)

// getGoroutineID extracts the runtime ID of the calling goroutine by parsing
// the first line of its stack trace.
func getGoroutineID() int64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.ParseInt(idField, 10, 64)
	return id
}
