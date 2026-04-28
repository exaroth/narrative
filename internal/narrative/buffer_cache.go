package narrative

import (
	"container/list"
	"sync"
)

// Value to be used when storing data in cache.
type BufferCacheVal struct {
	// Index of sentence for given source.
	SentenceIdx int
	// Waveform data.
	Data []float32
}

// Stores waveform data used for playback.
type BufferCacheLRU struct {
	l        *list.List
	elements map[int]*list.Element
	size     int
	capacity int
	mut      sync.Mutex
}

// NewLRU represent initiate lru cache with capacity
func NewBufferCacheLRU(capacity int) BufferCacheLRU {
	return BufferCacheLRU{
		l:        list.New(),
		elements: make(map[int]*list.Element, capacity),
		size:     0,
		capacity: capacity,
		mut:      sync.Mutex{},
	}
}

// Get value from BufferCacheLRU, or nil if not found.
func (c *BufferCacheLRU) Get(sentence_idx int) *BufferCacheVal {
	c.mut.Lock()
	defer c.mut.Unlock()
	v, ok := c.elements[sentence_idx]
	if ok {
		c.l.MoveToBack(v)
		return v.Value.(*BufferCacheVal)
	}

	return nil
}

// Insert new element into buffer cache, items already in cache will
// not be updated.
func (c *BufferCacheLRU) Put(sentence_idx int, data []float32) {
	c.mut.Lock()
	defer c.mut.Unlock()
	e, ok := c.elements[sentence_idx]
	if ok {
		// Don't need to update data, simply move to front. (kw)
		// if update {
		// 	n := e.Value.(*BufferCacheVal)
		// 	n.Data = data
		// 	e.Value = n
		// }
		c.l.MoveToBack(e)
		return
	}

	if c.size >= c.capacity {
		e := c.l.Front()
		idx := e.Value.(*BufferCacheVal).SentenceIdx
		c.l.Remove(e)
		delete(c.elements, idx)
		c.size--
	}

	n := BufferCacheVal{
		SentenceIdx: sentence_idx,
		Data:        data,
	}
	c.l.PushBack(n)
	elem := c.l.Back()
	c.elements[sentence_idx] = elem
	c.size++
}
