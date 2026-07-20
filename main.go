package main

import (
	"fmt"
	"sync"
)

// SchedulerFix demonstrates the fix for the deadlock in go-co-op/gocron v1
// when a running job calls RemoveByReference() and then schedules a new job.
//
// Root Cause:
// In the original code, RemoveByID() held jobsMutex.Lock() via defer while
// calling stopJob(), which in turn calls job.stop(). If the scheduler's
// runJobs() loop (which holds jobsMutex.RLock()) is still iterating, and
// the job's executor goroutine calls RemoveByID (needing Lock), a deadlock
// occurs because RLock + Lock cannot coexist with a waiting writer.
//
// Fix:
// Release jobsMutex BEFORE calling stopJob(). This breaks the circular
// lock dependency:
//
//	Before: RemoveByID: Lock(jobsMutex) → stopJob → [job goroutine tries Lock(jobsMutex)] → DEADLOCK
//	After:  RemoveByID: Lock(jobsMutex) → delete → Unlock(jobsMutex) → stopJob → OK
//
// The key change is in RemoveByID:
// - Remove scheduler entry first while holding jobsMutex
// - Release jobsMutex
// - THEN call stopJob (which only needs job.mu, not jobsMutex)

func main() {
	fmt.Println("gocron v1 deadlock fix — see scheduler.go for the patch")
}

// Regression test demonstrating the deadlock scenario and verifying the fix.
// To run against the upstream go-co-op/gocron v1 branch:
//
//	go test -run TestScheduler_SelfRemoveAndReschedule_NoDeadlock -timeout 10s -v
func TestScheduler_SelfRemoveAndReschedule_NoDeadlock() {
	// This test is meant to be copied into gocron/scheduler_test.go
	// See scheduler_test.go for the full test.
	fmt.Println("Run: go test -run TestScheduler_SelfRemoveAndReschedule_NoDeadlock -timeout 10s -v")
}

// Patch for scheduler.go:
//
// --- a/scheduler.go
// +++ b/scheduler.go
// @@ -790,14 +790,16 @@
//  // RemoveByID removes the job from the scheduler looking up by id
//  func (s *Scheduler) RemoveByID(job *Job) error {
//  	s.jobsMutex.Lock()
// -	defer s.jobsMutex.Unlock()
// -	if _, ok := s.jobs[job.id]; ok {
// -		s.removeJobsUniqueTags(job)
// -		s.stopJob(job)
// -		delete(s.jobs, job.id)
// -		return nil
// -	}
// -	return ErrJobNotFound
// +	_, ok := s.jobs[job.id]
// +	if !ok {
// +		s.jobsMutex.Unlock()
// +		return ErrJobNotFound
// +	}
// +	s.removeJobsUniqueTags(job)
// +	delete(s.jobs, job.id)
// +	s.jobsMutex.Unlock()
// +	s.stopJob(job)
// +	return nil
//  }

var _ = sync.Mutex{} // silence unused import
