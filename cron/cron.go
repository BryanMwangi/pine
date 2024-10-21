package cron

import (
	"sync"
	"time"

	"github.com/BryanMwangi/pine/logger"
	"github.com/google/uuid"
)

type Config struct {
	// When set to true the server will attempt to restart failed jobs
	RestartOnError bool

	// The number of times a job will be retried before it is deleted
	// when an error occurs
	//
	// Default: 0
	RetryAttempts int

	// This is the periodic time in which the server can execute
	// background tasks background tasks can run infinitely
	// as long as the server is running
	// for example you can use this to make requests to other servers
	// or update your database
	//
	// Default: 5 minutes
	BackgroundTimeout time.Duration
}

type Cron struct {
	// The configuration for the cron
	config Config

	// A slice of jobs that will be executed in the background
	jobs []Job

	// counts the number of retry attempts for each job
	retryCount map[uuid.UUID]int

	// Ensures that updates to the jobs slice are thread safe
	mutex sync.Mutex
}

// This is the structure of a background job
// you can use this to put whatever jobs you want to perform
// in the background as the server runs and Pine will take care of executing
// them in the background
//
// time is optional and defaults to 5 minutes according to the server configuration
//
// Fn is the function that will be executed.
// It should always return an error.
// If error is not nil the error will be used to delete the
// task from the queue otherwise when nil the task will run indefinitely
type Job struct {
	id   uuid.UUID
	Fn   func() error
	Time time.Duration
}

const (
	DefaultRetryAttempts  = 0
	DefaultRestartOnError = false
)

func New(cfg ...Config) *Cron {
	config := Config{
		RestartOnError:    DefaultRestartOnError,
		RetryAttempts:     DefaultRetryAttempts,
		BackgroundTimeout: 5 * time.Minute,
	}

	// We use the first config in the slice
	if len(cfg) > 0 {
		userConfig := cfg[0]

		if userConfig.RestartOnError {
			config.RestartOnError = userConfig.RestartOnError
		}
		if userConfig.RetryAttempts != 0 {
			config.RetryAttempts = userConfig.RetryAttempts
		}
		if userConfig.BackgroundTimeout != 0 {
			config.BackgroundTimeout = userConfig.BackgroundTimeout
		}
	}

	return &Cron{
		config:     config,
		retryCount: make(map[uuid.UUID]int),
	}
}

func (c *Cron) AddJobs(jobs ...Job) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var newJobs []Job
	for _, j := range jobs {
		j.id = uuid.New()
		newJobs = append(newJobs, j)
	}
	c.jobs = append(c.jobs, newJobs...)

}

func (c *Cron) removeJob(id uuid.UUID) {
	for i, j := range c.jobs {
		if j.id == id {
			c.jobs = append(c.jobs[:i], c.jobs[i+1:]...)
			break
		}
	}
	delete(c.retryCount, id)
}

func (c *Cron) handleJobError(job Job) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if the config has a restart policy
	// If no restart policy is set, we delete the job immediately
	if !c.config.RestartOnError {
		logger.RuntimeError("No restart policy set for job, deleting job...")
		c.removeJob(job.id)
		return
	}

	//we increment the retry count for the job
	c.retryCount[job.id]++

	//we check if the job has been retried the maximum number of times
	//if it has we delete it
	if c.config.RestartOnError && c.config.RetryAttempts > 0 && c.retryCount[job.id] >= c.config.RetryAttempts {
		logger.RuntimeError("Max retry attempts reached, deleting job")
		c.removeJob(job.id)
		return
	}

	logger.RuntimeInfo("Job will be retried in " + job.Time.String())
}

func (c *Cron) startJob(job Job) {
	for {
		// Execute the task function
		err := job.Fn()
		if err != nil {
			// Log the error
			logger.RuntimeError("Error in cron job")
			logger.RuntimeError(getFunctionName(job.Fn))
			logger.RuntimeError(err.Error())

			// Remove the task if it fails
			c.handleJobError(job)
			// If the job has been removed, exit the loop
			if !c.jobExists(job.id) {
				break
			}
		}
		// Respect the delay specified by the task
		if job.Time > 0 {
			time.Sleep(job.Time)
		} else {
			time.Sleep(c.config.BackgroundTimeout)
		}
	}
}

func (c *Cron) jobExists(id uuid.UUID) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, j := range c.jobs {
		if j.id == id {
			return true
		}
	}
	return false
}

// Call this method to start the cron
//
// By default cron jobs are executed in their own goroutines hence in separate threads
// This method starts the cron jobs in a new thread away from the server's main thread
//
// This ensures non blocking execution of cron jobs hence the server
// can handle requests as cron jobs are executed with minimal impact on the server
func (c *Cron) Start() {
	go c.processCron()
}

// Internal method used to start specific cron jobs
func (c *Cron) processCron() {
	for _, job := range c.jobs {
		go c.startJob(job) // Start the background task
	}
}
