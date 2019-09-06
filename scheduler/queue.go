package scheduler

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Task encapsulates a function that will be run at a specific time.
type Task interface {
	// When returns the time at which the task is to be performed.
	When() time.Time
	// Run is called when the task is performed.
	Run(time.Time)
}

// NewTask creates a task that calls run at when.
func NewTask(when time.Time, run func(time.Time)) Task {
	return &task{
		when: when,
		run:  run}
}

// task is a basic struct implementing the Task interface
// available for convenience.
type task struct {
	when time.Time
	run  func(time.Time)
}

func (t *task) When() time.Time { return t.when }

func (t *task) Run(when time.Time) { t.run(when) }

func (t *task) String() string {
	return t.When().String()
}

// TaskQueue is a concurrent-safe queue of tasks.
type TaskQueue struct {
	queue   []Task
	m       sync.Mutex
	running int64
	wg      sync.WaitGroup
}

// NewTaskQueue creates a new TaskQueue.
func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		queue: []Task{}}
}

// Append inserts a Task into the queue.
func (q *TaskQueue) Append(t Task) {
	q.m.Lock()
	defer q.m.Unlock()

	// find index for insertion
	var i int
	for ; i < len(q.queue) && q.queue[i].When().Before(t.When()); i++ {
	}

	q.queue = append(q.queue, nil)   // grow queue with zero value
	copy(q.queue[i+1:], q.queue[i:]) // shift contents up one index
	q.queue[i] = t                   // insert item
}

// Process calls Task.Run() on all tasks due.
func (q *TaskQueue) Process() {
	q.m.Lock()
	defer q.m.Unlock()

	now := time.Now()
	var front int
	for i := 0; i < len(q.queue); i++ {
		if q.queue[i].When().Before(now) { // TODO: what about tasks older than a certain duration?

			q.wg.Add(1)
			atomic.AddInt64(&q.running, 1)
			go func(t Task) {
				t.Run(t.When())
				q.wg.Done()
				atomic.AddInt64(&q.running, -1)
			}(q.queue[i])

			q.queue[i] = nil
			front = i + 1
		} else {
			break
		}
	}

	// move all remaining tasks forward
	keep := q.queue[front:]
	for i := range keep {
		q.queue[i] = keep[i]
	}

	// slice off the 'extra'
	q.queue = q.queue[:len(keep)]
}

// Next gets the time the next Task must be processed.
func (q *TaskQueue) Next() time.Time {

	if q.Len() > 0 {
		q.m.Lock()
		defer q.m.Unlock()
		return q.queue[0].When()
	}
	return time.Now()
}

// Len returns the number of Tasks currently in the queue.
func (q *TaskQueue) Len() int {
	q.m.Lock()
	defer q.m.Unlock()
	return len(q.queue)
}

// Running returns the number of Tasks currently running.
func (q *TaskQueue) Running() int {
	q.m.Lock()
	defer q.m.Unlock()
	return int(q.running)
}

// Wg gives access to a WaitGroup for the running tasks.
func (q *TaskQueue) Wg() *sync.WaitGroup {
	return &q.wg
}

func (q *TaskQueue) String() string {
	// q.m.Lock()
	// defer q.m.Unlock()
	return fmt.Sprintf("%d in queue| %d running| next task at %s", q.Len(), q.Running(), q.Next())
}
