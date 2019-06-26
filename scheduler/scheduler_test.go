package scheduler

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TODO: this is a 'fake' test. I'm not really sure how to
// test scheduler.
func TestNewScheduler(t *testing.T) {

	f := func(num int) func(time.Time) {
		return func(t time.Time) {
			fmt.Printf("%-3d%s\n", num, t)
		}
	}

	sch := NewScheduler()
	sch.Add(NewTask(time.Now().Add(30*time.Second), f(1)))

	ctx, _ := context.WithTimeout(context.Background(), 35*time.Second)
	sch.Start(ctx)

	for i := 2; i < 6; i++ {
		sch.Add(NewTask(time.Now().Add(time.Duration(i*5)*time.Second), f(i)))
	}

	sch.Wait()
	// time.Sleep(10 * time.Millisecond)
	fmt.Println("queue is empty. done.")
}
