package main

import (
	"fmt"
	"sync"
)

// Scheduler with deadlock-free concurrent job removal and rescheduling.
// The bug: concurrent Remove() and Reschedule() deadlock.
// Fix: Use a single lock and never hold it during callback execution.
type Scheduler struct {
	mu   sync.Mutex
	jobs map[string]bool
}

func NewScheduler() *Scheduler {
	return &Scheduler{jobs: make(map[string]bool)}
}

func (s *Scheduler) Add(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[name] = true
}

func (s *Scheduler) Remove(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.jobs[name] {
		delete(s.jobs, name)
		return true
	}
	return false
}

func (s *Scheduler) Reschedule(name string) bool {
	removed := s.Remove(name)
	if removed {
		s.Add(name)
	}
	return removed
}

func main() {
	s := NewScheduler()
	s.Add("job1")
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); s.Reschedule("job1") }()
	}
	wg.Wait()
	fmt.Println("No deadlock - all reschedule operations completed")
}
