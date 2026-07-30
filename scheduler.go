package gocron

import "sync"

type Scheduler struct {
	mu   sync.Mutex
	jobs map[string]*Job
}

type Job struct {
	id string
}

func (s *Scheduler) RemoveAndReschedule(id string, newJob *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, id)
	s.jobs[newJob.id] = newJob
}
