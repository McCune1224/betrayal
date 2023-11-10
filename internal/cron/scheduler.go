package cron

import (
	"time"

	"github.com/go-co-op/gocron"
)

// BetrayalScheduler is a wrapper around gocron.Scheduler
// that adds additional functionality to the scheduler
// for the Betrayal Discord bot.
type BetrayalScheduler struct {
	s     *gocron.Scheduler
	tasks map[string]*gocron.Job
}

func LoadScheduler() *BetrayalScheduler {
	return &BetrayalScheduler{
		s: gocron.NewScheduler(time.UTC),
	}
}

func (bs *BetrayalScheduler) GetTasks() map[string]*gocron.Job {
	return bs.tasks
}

func (bs *BetrayalScheduler) GetScheduler() *gocron.Scheduler {
	return bs.s
}

func (bs *BetrayalScheduler) Start() {
	bs.s.StartAsync()
}

func (bs *BetrayalScheduler) Stop() {
	bs.s.Stop()
}

func (bs *BetrayalScheduler) Clear() {
	bs.s.Clear()
}
