package scheduler

import (
	"context"
	"sync"
	"time"
)

type Scheduler struct {
	queue            *safeTaskQueue
	timer            *time.Timer
	done             chan struct{}
	stopOnEmptyQueue bool
	mutex            sync.Mutex
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		queue: newSafeQueue(),
		timer: time.NewTimer(-1)}
}

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
					// fmt.Println("EMPTY QUEUE")
					close(s.done)
					return
				}

			case <-ctx.Done():
				// fmt.Println("EXITING SCHEDULER LOOP")
				close(s.done)
				return
			}
		}
	}()
}

func (s *Scheduler) Wait() {
	<-s.done
}

func (s *Scheduler) WaitUntilQueueEmpty() {
	s.stopOnEmptyQueue = true
	s.Wait()
}

func (s *Scheduler) resetTimer(t time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !s.timer.Stop() { // stupid bullshit
		select {
		case <-s.timer.C:
			// fmt.Println("DRAINED TIMER CHANNEL")
		default:
		}
	}
	s.timer.Reset(time.Until(t))
}

func (s *Scheduler) Add(t Task) {
	s.queue.Append(t)
	s.resetTimer(s.queue.Next())
}

func (s *Scheduler) String() string {
	return s.queue.String()
}
