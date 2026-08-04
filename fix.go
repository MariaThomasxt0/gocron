package fix

import "sync"

// FixDeadlock uses separate locks for different operations to prevent
// deadlocks during concurrent job removal and rescheduling.
type Scheduler struct {
    jobsMu     sync.RWMutex
    scheduleMu sync.Mutex
}

func (s *Scheduler) RemoveJob(id string) {
    s.jobsMu.Lock()
    defer s.jobsMu.Unlock()
    // remove job logic
}

func (s *Scheduler) Reschedule(id string) {
    s.scheduleMu.Lock()
    defer s.scheduleMu.Unlock()
    // reschedule logic
}
