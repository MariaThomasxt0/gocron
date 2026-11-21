# Bounty Fix

**Reward:** $10.00
**Issue:** https://github.com/MariaThomasxt0/gocron/issues/1
**Repo:** https://github.com/MariaThomasxt0/gocron

---

## Issue

🎯 Fix Deadlock on Concurrent Job Removal and Rescheduling

## 📝 Description
A deadlock occurs in the scheduler when a running job attempts to dynamically modify the scheduler's state by removing itself (or another job) and immediately scheduling a replacement job. 

Under specific timing conditions, the scheduler's main execution loop and the goroutine running the job enter a deadlock state. This happens because:
1. The job execution goroutine calls `RemoveByReference()` (or a similar removal API), which acquires the scheduler's internal lock.
2. Before the lock is fully released or the state is reconciled, the same job goroutine attempts to schedule a new job (e.g., calling `.Every().Do(...)`), which requests the scheduler lock again.
3. Concurrently, the scheduler's main execution loop attempts to update the job registry or check job states, leading to circular lock contention.

As a result, both the scheduler loop and the job execution goroutine block indefinitely, halting all subsequent job executions.

## 🎯 Acceptance Criteria
- [ ] A running job must be able to safely call `RemoveByReference()` (or equivalent) and schedule a new job sequentially within its execution block without causing a deadlock.
- [ ] The scheduler's internal job registry must remain consistent and thread-safe during concurrent read, write, and delete operations.
- [ ] The scheduler's main execution loop must not block indefinitely if a job modifies the scheduler state during its execution.
- [ ] A regression test must be implemented that reproduces this specific sequence (Execute -> Remove -> Schedule) and verifies it completes successfully without hanging.

## 🛠️ Technical Specifications & Context
- **Target Files:** Likely `scheduler.go`, `job.go`, and associated internal registry structures.
- **Locking Mechanism:** The `Scheduler` struct typically uses a `sync.Mutex` or `sync.RWMutex` to protect the slice/map of jobs. 
- **The Root Cause:** Nested lock acquisition or holding the scheduler lock while waiting for channel operations/goroutine synchronization. Specifically, if `RemoveByReference` and the job creation functions both lock the scheduler, and the scheduler's run loop is waiting on a channel or lock while holding another, a deadlock occurs.
- **Proposed Solutions:**
  1. **Lock Refactoring:** Ensure that locks are released before invoking any channel sends or operations that might block. Avoid holding the scheduler lock across execution boundaries.
  2. **Internal Event Queue / Channel-based Updates:** Instead of directly modifying the job slice and acquiring locks from the job's goroutine, queue the "Remove" and "Add" operations onto an internal control channel that the scheduler's main loop processes sequentially. This eliminates lock contention between the execution loop and job goroutines.

## 🧪 Verification & Testing
Create a test case in `scheduler_test.go` (or a new test file) that mimics the deadlock scenario:

```go
func TestScheduler_SelfRemoveAndReschedule_NoDeadlock(t *testing.T) {
    s := NewScheduler(time.UTC)
    
    var wg sync.WaitGroup
    wg.Add(1)

    // Schedule a job that will remove itself and schedule a new one
    var job *Job
    var err error
    job, err = s.Every(10 * time.Millisecond).Do(func() {
        // 1. Remove self
        s.RemoveByReference(job)
        
        // 2. Schedule a new job immediately
        _, err2 := s.Every(10 * time.Millisecond).Do(func() {
            wg.Done()
        })
        if err2 != nil {
            t.Errorf("failed to schedule new job: %v", err2)
        }
    })
    if err != nil {
        t.Fatalf("failed to schedule initial job: %v", err)
    }

    s.StartAsync()
    defer s.Stop()

    // Use a channel with a timeout to detect deadlocks
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        // Success
    case <-time.After(2 * time.Second):
        t.Fatal("Test timed out: Deadlock detected during self-removal and rescheduling")
    }
}
```

---

[![Opire Bounty](https://img.shields.io/badge/bounty-Opire-emerald?style=for-the-badge)](https://github.com/MariaThomasxt0/gocron)
<br/>
<hr/>

<details><summary>This repo is using Opire - what does it mean? 👇</summary><br/>💵 Everyone can add rewards for this issue commenting <code>/reward 100</code> (replace <code>100</code> with the amount).<br/>🕵️‍♂️ If someone starts working on this issue to earn the rewards, they can comment <code>/try</code> to let everyone know!<br/>🙌 And when they open the PR, they can comment <code>/claim #1</code> either in the PR description or in a PR's comment.<br/><br/>🪙 Also, everyone can tip any user commenting <code>/tip 20 @MariaThomasxt0</code> (replace <code>20</code> with the amount, and <code>@MariaThomasxt0</code> with the user to tip).<br/><br/>📖 If you want to learn more, check out our <a href="https://docs.opire.dev">documentation</a>.</details>

---

## Process

### 1. Orient yourself first
Before touching any code:
- Read the README and any relevant docs
- Run: find . -type f | grep -v node_modules | grep -v .git | grep -v dist | head -60
- Grep for the key terms from the issue title to find the relevant files
- Read the files around the bug — understand what's currently happening

### 2. Find and read the tests
- Look for test files near the relevant code (*.test.ts, *.spec.js, __tests__, etc.)
- Read them — they tell you what behavior is expected
- Understand what a passing fix looks like before writing a line

### 3. Implement the fix
- Make the MINIMAL change that fixes the issue — nothing more
- Do not refactor, rename, reorganize, or clean up unrelated code
- Do not add dependencies unless the issue explicitly requires it
- Do not modify config files (package.json, tsconfig.json, etc.) unless that IS the fix
- One logical change, one commit

### 4. Verify
- Run the test suite: check package.json scripts for "test", "lint", "typecheck"
  Common commands: npm test, yarn test, pnpm test, pytest, go test ./..., cargo test
- If TypeScript is present: npx tsc --noEmit
- Fix any test failures your change caused — if a test was wrong before, update it

### 5. Commit and document
- git add -A
- git commit -m "fix: 🎯 Fix Deadlock on Concurrent Job Removal and Rescheduling"
- Write AGENT_SUMMARY.md with:
  - Root cause: what was wrong and why
  - Fix: what you changed and the reasoning
  - Verification: how to confirm it works
  - Concerns: anything a reviewer should know

---

## Stop conditions

If ANY of these are true, write "UNCERTAIN: [reason]" at the top of AGENT_SUMMARY.md, make NO code changes, and stop:
- The fix requires changing more than ~5 files
- You cannot determine what the correct behavior should be
- The issue requires a database migration, API key, or external service
- You find 3+ open PRs already attempting this fix (high competition)
- The repo is not in a language/framework you can reason about confidently

A clean UNCERTAIN is better than a bad PR.
