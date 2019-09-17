// package scheduler implements a task scheduler.

package scheduler

import (
	"context"
	"sync"
	"time"
)

// Option modifies the Scheduler.
type Option func(*Scheduler)

// WaitForUnfinishedTasks configures a Scheduler to wait up to d time
// for unfinished Tasks to complete when the context provided to
// Start() is done.
func WaitForUnfinishedTasks(d time.Duration) Option {
	return func(s *Scheduler) {
		s.waitTimeout = d
	}
}

// StopWhenQueueEmpty configures a Scheduler to stop running when there
// are no more Tasks in its queue.
func StopWhenQueueEmpty() Option {
	return func(s *Scheduler) {
		s.stopOnEmptyQueue = true
	}
}

// Scheduler keeps a queue of tasks and processes them at their scheduled time.
type Scheduler struct {

	// primary processing mechanisms
	queue *TaskQueue
	timer *time.Timer

	// config options
	stopOnEmptyQueue bool
	waitTimeout      time.Duration

	// concurrency control
	done  chan struct{}
	mutex sync.Mutex
}

// NewScheduler creates a new Scheduler for use with the provided options.
func NewScheduler(options ...Option) *Scheduler {
	s := &Scheduler{
		queue: NewTaskQueue(),
		timer: time.NewTimer(-1)}

	// apply options
	for _, opt := range options {
		opt(s)
	}

	return s
}

// Start begins the scheduler process using the given context.
func (s *Scheduler) Start(ctx context.Context) {
	s.done = make(chan struct{})

	go func() {
		for {
			select {

			case <-s.timer.C:
				if s.queue.Len() > 0 {
					s.queue.Process()
					s.resetTimer(s.queue.Next())
				} else if s.stopOnEmptyQueue {
					// the queue is empty and the scheduler is configured to
					// stop processing tasks once the queue is empty.
					close(s.done)
					return
				}

			case <-ctx.Done():
				// the context was canceled or whatever.
				// this will wait up to waitTimeout for the tasks currently
				// running finish.
				tasksDone := make(chan struct{})
				go func() {
					defer close(tasksDone)
					s.queue.Wg().Wait()
				}()

				select {
				case <-time.After(s.waitTimeout):
				case <-tasksDone:
				}

				close(s.done)
				return
			}
		}
	}()
}

// resets the timer.
func (s *Scheduler) resetTimer(t time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !s.timer.Stop() { // stupid bullshit
		select {
		case <-s.timer.C:
			// draining timer channel per docs.
		default:
		}
	}
	s.timer.Reset(time.Until(t))
}

// Wait blocks the current goroutine until the context passed to Start()
// is canceled PLUS any amount of "wait time" the scheduler was configured with.
func (s *Scheduler) Wait() {
	<-s.done
}

// Add enqueues a task.
func (s *Scheduler) Add(t Task) {
	s.queue.Append(t)
	s.resetTimer(s.queue.Next())
}

// Running returns the number of Tasks currently running.
func (s *Scheduler) Running() int {
	return s.queue.Running()
}

func (s *Scheduler) String() string {
	return s.queue.String()
}
