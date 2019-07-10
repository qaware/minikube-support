package logs

import "sync"

const size uint64 = 256
const indexMask uint64 = size - 1

// buffer is a simple ring buffer which provides a capacity of size and evicts
// the oldest entries if new one are stored.
type buffer struct {
	entries            [size]interface{}
	lastCommittedIndex uint64
	nextFreeIndex      uint64
	mu                 sync.Mutex
}

// newBuffer creates a new ring buffer.
func newBuffer() *buffer {
	return &buffer{
		entries:            [size]interface{}{},
		lastCommittedIndex: 0,
		nextFreeIndex:      0,
		mu:                 sync.Mutex{},
	}
}

// Write adds a new entry to the ring buffer. It may evict the oldest entry if
// the buffer is full.
func (b *buffer) Write(value interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	currentIndex := b.nextFreeIndex
	b.nextFreeIndex = (b.nextFreeIndex + 1) & indexMask

	b.entries[currentIndex&indexMask] = value
	b.lastCommittedIndex = currentIndex & indexMask
}

// GetEntries returns a list of all entries. The oldest entry will be the first
// one in the returned list. While the newest entry is the last entry in the returned list.
func (b *buffer) GetEntries() []interface{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	var result = make([]interface{}, size)

	index := b.nextFreeIndex
	for i := uint64(0); i < size; i++ {
		result[i] = b.entries[index&indexMask]

		index++
	}
	return result
}
