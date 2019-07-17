package scheduler

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Task encapsulates a function that will be run at a specific time.
type Task interface {
	// When returns the time at which the task is to be performed.
	When() time.Time
	// Run is called when the task is performed.
	Run(time.Time)
}

type task struct {
	when time.Time
	run  func(time.Time)
}

// NewTask creates a task that calls run at when.
func NewTask(when time.Time, run func(time.Time)) Task {
	return &task{
		when: when,
		run:  run}
}

func (t *task) When() time.Time { return t.when }

func (t *task) Run(when time.Time) { t.run(when) }

func (t *task) String() string {
	return t.When().String()
}

// taskQueue is a queue of tasks to be executed.
type taskQueue []Task

func (q taskQueue) Append(t Task) taskQueue {
	// find index for insertion
	var i int
	for ; i < len(q) && q[i].When().Before(t.When()); i++ {
	}

	q = append(q, nil)   // grow queue with zero value
	copy(q[i+1:], q[i:]) // shift contents up one index
	q[i] = t             // insert item

	return q
}

func (q taskQueue) sort() {
	sort.SliceStable(q, func(i, j int) bool {
		return q[i].When().Before(q[j].When())
	})
}

func (q taskQueue) Process() taskQueue {
	// q.sort()

	now := time.Now()
	var front int
	for i := 0; i < len(q); i++ {
		if q[i].When().Before(now) {
			go q[i].Run(q[i].When())
			q[i] = nil
			front = i + 1
		} else {
			break
		}
	}

	keep := q[front:]
	for i := range keep {
		q[i] = keep[i]
	}

	return q[:len(keep)]
}

// safeTaskQueue is a concurrent-safe taskQueue.
type safeTaskQueue struct {
	sync.Mutex
	queue taskQueue
}

func newSafeQueue() *safeTaskQueue {
	return &safeTaskQueue{
		queue: taskQueue{}}
}

func (q *safeTaskQueue) Append(t Task) {
	q.Lock()
	defer q.Unlock()
	q.queue = q.queue.Append(t)
}

func (q *safeTaskQueue) Process() {
	q.Lock()
	defer q.Unlock()
	q.queue = q.queue.Process()
}

func (q *safeTaskQueue) Next() time.Time {

	if q.Len() > 0 {
		q.Lock()
		defer q.Unlock()
		return q.queue[0].When()
	}
	return time.Now()
}

func (q *safeTaskQueue) Len() int {
	q.Lock()
	defer q.Unlock()
	return len(q.queue)
}

func (q *safeTaskQueue) String() string {
	q.Lock()
	defer q.Unlock()
	return fmt.Sprintf("len(%d) %v", len(q.queue), q.queue)
}
