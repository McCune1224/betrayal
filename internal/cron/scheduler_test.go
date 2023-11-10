package cron

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/go-co-op/gocron"
)

// Random Test to see if I can get this to work in the sandbox
func TestScheduleOneTimeTask(t *testing.T) {
	s := gocron.NewScheduler(time.UTC)

	// Schedule a one-time task to run after 2 minutes
	_, _ = s.Every(5).Second().WaitForSchedule().LimitRunsTo(1).Do(func() {
		fmt.Println("One-Time Task executed at:", time.Now())
	})

	// Start the scheduler
	s.StartAsync()

	// Run the scheduler for a certain duration (e.g., 3 minutes)
	time.Sleep(3 * time.Minute)

	// Stop the scheduler
	s.Stop()
}

func task() {
	log.Println("I am a task")
}

func taskWithParams(a int, b string) {
	log.Println(a, b)
}
