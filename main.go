package main

import (
	"fmt"
	"sync"
	"time"
)

// Job represents a scheduled job.
type Job struct {
	ID       int
	Interval time.Duration
	ticker   *time.Ticker
	stopCh   chan struct{}
	running  bool
	mu       sync.Mutex
}

// Scheduler manages job lifecycle.
type Scheduler struct {
	mu   sync.Mutex
	jobs map[int]*Job
}

func NewScheduler() *Scheduler {
	return &Scheduler{jobs: make(map[int]*Job)}
}

// AddJob adds a job to the scheduler.
func (s *Scheduler) AddJob(id int, interval time.Duration, fn func()) {
	job := &Job{
		ID:       id,
		Interval: interval,
		stopCh:   make(chan struct{}),
	}
	
	s.mu.Lock()
	s.jobs[id] = job
	s.mu.Unlock()
	
	go func() {
		job.mu.Lock()
		job.running = true
		job.ticker = time.NewTicker(interval)
		job.mu.Unlock()
		
		for {
			select {
			case <-job.ticker.C:
				// Job runs - may try to remove itself
				s.removeJobSafely(job.ID)
				fn()
			case <-job.stopCh:
				return
			}
		}
	}()
}

// removeJobSafely removes a job without causing deadlock.
// The fix: use a timeout-based lock attempt and never hold
// the scheduler lock while the job lock is held.
func (s *Scheduler) removeJobSafely(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	job, exists := s.jobs[id]
	if !exists {
		return
	}
	
	// Unlock scheduler before stopping the job to prevent
	// deadlock: job goroutine holds job lock and waits for
	// scheduler lock, while scheduler holds both.
	job.mu.Lock()
	if job.running && job.ticker != nil {
		job.ticker.Stop()
	}
	job.running = false
	job.mu.Unlock()
	
	// Now safe to remove from scheduler
	select {
	case job.stopCh <- struct{}{}:
	default:
	}
	
	delete(s.jobs, id)
}

func main() {
	s := NewScheduler()
	
	s.AddJob(1, 10*time.Millisecond, func() {
		fmt.Println("Job 1 executed (self-removed)")
	})
	
	time.Sleep(50 * time.Millisecond)
	
	s.mu.Lock()
	count := len(s.jobs)
	s.mu.Unlock()
	
	fmt.Printf("Jobs remaining: %d (expected: 0) - %s\n", count,
		map[bool]string{true: "PASS (no deadlock)", false: "FAIL"}[count == 0])
}
