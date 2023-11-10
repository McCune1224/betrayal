package cron

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
)

// BetrayalScheduler is a wrapper around gocron.Scheduler
// that adds additional functionality to the scheduler
// for the Betrayal Discord bot.
type BetrayalScheduler struct {
	// internal gocron scheduler
	s *gocron.Scheduler
	// internal map of jobIDs to jobs for easy access and modification
	jobs map[string]*gocron.Job
}

func NewScheduler() *BetrayalScheduler {
	return &BetrayalScheduler{
		s:    gocron.NewScheduler(time.UTC),
		jobs: make(map[string]*gocron.Job),
	}
}

func (bs *BetrayalScheduler) GetJobs() map[string]*gocron.Job {
	bs.cleanup()
	return bs.jobs
}

func (bs *BetrayalScheduler) GetJob(jobID string) (*gocron.Job, error) {
	bs.cleanup()
	if job, ok := bs.jobs[jobID]; ok {
		return job, nil
	}
	return nil, fmt.Errorf("job %s not found", jobID)
}

// Insert a one-time job into the scheduler, will overwrite any existing job with the same ID
func (bs *BetrayalScheduler) UpsertJob(jID string, dur time.Duration, jf interface{}) error {
	bs.cleanup()
	// If the jobID already exists, replace the existing job
	// with the new job

	job, err := bs.s.Every(dur).WaitForSchedule().LimitRunsTo(1).Do(jf)
	if err != nil {
		return err
	}

	if _, ok := bs.jobs[jID]; ok {
		bs.s.Remove(bs.jobs[jID])
		delete(bs.jobs, jID)
	}
	bs.jobs[jID] = job
	return nil
}

func (bs *BetrayalScheduler) DeleteJob(jobID string) error {
	bs.cleanup()
	if job, ok := bs.jobs[jobID]; ok {
		bs.s.Remove(job)
		delete(bs.jobs, jobID)
	}
	return nil
}

// Get uunderlying gocron.Scheduler
func (bs *BetrayalScheduler) GetScheduler() *gocron.Scheduler {
	return bs.s
}

// Start the scheduler
func (bs *BetrayalScheduler) Start() {
	bs.cleanup()
	bs.s.StartAsync()
}

func (bs *BetrayalScheduler) Restart() {
	bs.Stop()
	bs.Start()
}

func (bs *BetrayalScheduler) Stop() {
	bs.jobs = make(map[string]*gocron.Job)
	bs.s.Stop()
}

func (bs *BetrayalScheduler) Clear() {
	bs.s.Clear()
	bs.jobs = make(map[string]*gocron.Job)
}

// WARNING: I really don't want to even consider dealing with async
// issues here, so I'm just going to do a cleanup every time
// we access anything in the scheduler
func (bs *BetrayalScheduler) cleanup() {
	now := time.Now()
	for jobID, job := range bs.jobs {
		if job.NextRun().Before(now) || job.FinishedRunCount() > 0 {
			// job is in the past, so remove it
			bs.s.Remove(job)
			delete(bs.jobs, jobID)
		}
	}
}
