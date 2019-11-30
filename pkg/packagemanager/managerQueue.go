package packagemanager

// item stores the current manager including the priority.
type item struct {
	manager  osSpecific
	priority int
	index    int
}

// queue contains all manager items to find the manager with the highest priority.
// For details about the implementation see the golang documentation under
// https://golang.org/pkg/container/heap/#example__priorityQueue
type queue []*item

func (q queue) Len() int { return len(q) }

func (q queue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return q[i].priority < q[j].priority
}

func (q queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *queue) Push(x interface{}) {
	n := len(*q)
	item := x.(*item)
	item.index = n
	*q = append(*q, item)
}

func (q *queue) Pop() interface{} {
	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*q = old[0 : n-1]
	return item
}
