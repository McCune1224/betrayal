package scheduler

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/mccune1224/betrayal/pkg/data"
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
	return nil, fmt.Errorf("job %s not found", jobID)
}

func (bs *BetrayalScheduler) JobExists(jobID string) bool {
	jobs, err := bs.s.FindJobsByTag(jobID)
	if err != nil {
		return false
	}
	if len(jobs) > 0 {
		return true
	}
	return false
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
	jobID := jobData.MakeJobID()

	invokeTime := time.Unix(jobData.InvokeTime, 0)
	if invokeTime.Before(time.Now()) {
		return ErrJobExpired
	}
	_, err := bs.s.Every(1).StartAt(invokeTime).WaitForSchedule().LimitRunsTo(1).Tag(jobID).Do(jf)
	if err != nil {
		return err
	}

	// check if the job already exists
	// if it does and it's not expired, skip DB insert
	// if it does and it is expired, delete it and insert the new one
	existing, err := bs.m.InventoryCronJobs.GetByJobID(jobID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if existing == nil {
		err = bs.m.InventoryCronJobs.Insert(jobData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bs *BetrayalScheduler) RescheduleJob(jobData *data.InventoryCronJob, jf func()) error {
	jobID := jobData.MakeJobID()
	job, err := bs.s.FindJobsByTag(jobID)
	if err != nil {
		return err
	}
	if len(job) > 1 {
		return ErrJobDuplicate
	}

	err = bs.s.RemoveByTag(jobID)
	if err != nil {
		return err
	}
	_, err = bs.s.Every(1).StartAt(time.Now()).WaitForSchedule().LimitRunsTo(1).Tag(jobID).Do(jf)
	if err != nil {
		return err
	}
	return nil
}

// Manualyly invoke a job by ID and then remove it from the database (really for when the job is already expired i.e bot down when timer expired)
func (bs *BetrayalScheduler) InvokeJob(jobID string, jf func()) error {
	_, err := bs.s.Every(1).StartAt(time.Now()).WaitForSchedule().LimitRunsTo(1).Do(jf)
	if err != nil {
		return err
	}
	err = bs.m.InventoryCronJobs.DeletebyJobID(jobID)
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
	bs.s.StartAsync()
}

func (bs *BetrayalScheduler) Stop() {
	bs.s.Stop()
}

func (bs *BetrayalScheduler) Clear() {
	bs.s.Clear()
}

// func (bs *BetrayalScheduler) cleanup() error {
// 	now := time.Now()
// 	tags := bs.s.GetAllTags()
// 	for _, t := range tags {
// 		jobs, err := bs.s.FindJobsByTag(t)
// 		if err != nil {
// 			return err
// 		}
// 		if len(jobs) > 1 {
// 			return ErrJobDuplicate
// 		}
//     log.Println(jobs[0].NextRun(), now)
//     log.Println(jobs[0].NextRun().Before(now))
// 		// if jobs[0].NextRun().Before(now) {
// 		// 	err := bs.DeleteJob(t)
// 		// 	if err != nil {
// 		// 		return err
// 		// 	}
// 		// }
// 	}
// 	return nil
// }
