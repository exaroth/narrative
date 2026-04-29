package narrative

import (
	"maps"
	"slices"
	"sync"
)

// Implements simple lock to make sure we never process
// 2 sentences at the same time.
type BufferCacheLock struct {
	procs map[int]struct{}
	mut   sync.Mutex
}

func (l *BufferCacheLock) Add(n int) {
	l.mut.Lock()
	defer l.mut.Unlock()
	l.procs[n] = struct{}{}
}

func (l *BufferCacheLock) Rm(n int) {
	l.mut.Lock()
	defer l.mut.Unlock()
	delete(l.procs, n)
}

func (l *BufferCacheLock) Has(n int) bool {
	l.mut.Lock()
	defer l.mut.Unlock()
	if _, ok := l.procs[n]; ok {
		return true
	}
	return false
}

func (l *BufferCacheLock) Items() []int {
	l.mut.Lock()
	defer l.mut.Unlock()
	return slices.Sorted(maps.Keys(l.procs))
}
