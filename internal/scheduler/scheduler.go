package scheduler

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/mccune1224/betrayal/internal/data"
)

var (
	ErrJobNotFound      = fmt.Errorf("job not found")
	ErrJobExpired       = fmt.Errorf("job expired")
	ErrJobAlreadyExists = fmt.Errorf("job already exists")
	ErrJobFailed        = fmt.Errorf("job failed")
	ErrJobDuplicate     = fmt.Errorf("job duplicate(s) found")
)

// BetrayalScheduler is a wrapper around gocron.Scheduler
// that adds additional functionality to the scheduler
// for the Betrayal Discord bot.
type BetrayalScheduler struct {
	// internal gocron scheduler
	m data.Models
	s *gocron.Scheduler
	// internal map of jobIDs to jobs for easy access and modification
}

func NewScheduler(dbJobs data.Models) *BetrayalScheduler {
	return &BetrayalScheduler{
		s: gocron.NewScheduler(time.UTC),
		m: dbJobs,
	}
}

func (bs *BetrayalScheduler) GetJob(jobID string) (*gocron.Job, error) {
	bs.cleanup()
	return nil, fmt.Errorf("job %s not found", jobID)
}

func (bs *BetrayalScheduler) DeleteJob(jobID string) error {
	jobs, err := bs.s.FindJobsByTag(jobID)
	if err != nil {
		return err
	}
	if len(jobs) > 1 {
		return ErrJobDuplicate
	}
	err = bs.s.RemoveByTag(jobID)
	if err != nil {
		return err
	}

	err = bs.m.InventoryCronJobs.DeletebyJobID(jobID)
	if err != nil {
		return err
	}
	return nil
}

// Insert a one-time job into the scheduler, will overwrite any existing job with the same ID
func (bs *BetrayalScheduler) InsertJob(jobData *data.InventoryCronJob, jf func()) error {
	bs.cleanup()
	jobID := jobData.MakeJobID()

	// check to make sure job isn't past due date
	if time.Unix(jobData.InvokeTime, 0).Before(time.Unix(jobData.StartTime, 0)) {
		bs.m.InventoryCronJobs.DeletebyJobID(jobID)
		return ErrJobExpired
	}

	dur := time.Duration(time.Until(time.Unix(jobData.InvokeTime, 0)).Seconds() * float64(time.Second))
	newJob, err := bs.s.Every(dur).WaitForSchedule().LimitRunsTo(1).Do(jf)
	if err != nil {
		return err
	}
	newJob.Tag(jobID)
	err = bs.m.InventoryCronJobs.Insert(jobData)
	if err != nil {
		return err
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

func (bs *BetrayalScheduler) Stop() {
	bs.s.Stop()
}

func (bs *BetrayalScheduler) Clear() {
	bs.s.Clear()
}

// WARNING: I really don't want to even consider dealing with async
// issues here, so I'm just going to do a cleanup every time
// we access anything in the scheduler
func (bs *BetrayalScheduler) cleanup() error {
	now := time.Now()
	tags := bs.s.GetAllTags()
	for _, t := range tags {
		jobs, err := bs.s.FindJobsByTag(t)
		if err != nil {
			return err
		}
		if len(jobs) > 1 {
			return ErrJobDuplicate
		}
		if jobs[0].NextRun().Before(now) {
			err := bs.DeleteJob(t)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
