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
			time.Sleep(10 * time.Second)
			fmt.Printf("%d done\n", num)
		}
	}

	fmt.Println("start:", nowplus(0))

	sch := NewScheduler(WaitForUnfinishedTasks(10 * time.Second))
	sch.Add(NewTask(nowplus(30), f(1)))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	sch.Start(ctx)

	for i := 2; i < 6; i++ {
		sch.Add(NewTask(nowplus(i*5), f(i)))
	}

	sch.Wait()
	// time.Sleep(10 * time.Millisecond)
	fmt.Println("15+10 sec since start. done.")
	fmt.Println(sch)
}

func nowplus(sec int) time.Time {
	n := time.Now().Add(time.Duration(sec) * time.Second)
	return time.Date(n.Year(), n.Month(), n.Day(), n.Hour(), n.Minute(), n.Second(), 0, n.Location())
}
