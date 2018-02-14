package concur_test

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stevenmatthewt/concur"
)

type MockJob struct {
	invokedUnatomic   int
	invoked           int32
	sleepDuration     time.Duration
	returnError       error
	returnValue       string
	actualReturnValue string
}

func (job *MockJob) Exec() error {
	if job.sleepDuration > 0 {
		time.Sleep(job.sleepDuration)
	}
	job.invokedUnatomic++
	atomic.AddInt32(&job.invoked, 1)
	job.actualReturnValue = job.returnValue
	return job.returnError
}

type PanicJob struct{}

func (job *PanicJob) Exec() error {
	panic("PanicJob panicking on purpose")
}

func TestConcurrentRunnerSimple(t *testing.T) {
	jobs := []MockJob{
		MockJob{},
		MockJob{},
		MockJob{},
		MockJob{},
		MockJob{},
		MockJob{},
		MockJob{},
		MockJob{},
		MockJob{},
	}

	err := concur.Concurrent().Run(&jobs[0], &jobs[1], &jobs[2], &jobs[3], &jobs[4], &jobs[5], &jobs[6], &jobs[7], &jobs[8])
	if err != nil {
		t.Error(err)
	}

	for i, job := range jobs {
		if job.invoked != 1 {
			t.Errorf("jobs[%d] invoked incorrect number of times: %d", i, job.invoked)
		}
	}
}

func TestConcurrentRunnerUneven(t *testing.T) {
	job1 := MockJob{}
	job2 := MockJob{}
	job3 := MockJob{}

	err := concur.Concurrent().Run(&job1, &job2, &job2, &job3, &job3, &job3)
	if err != nil {
		t.Error(err)
	}

	if job1.invoked != 1 {
		t.Errorf("job1 invoked incorrect number of times: %d", job1.invoked)
	}
	if job2.invoked != 2 {
		t.Errorf("job2 invoked incorrect number of times: %d", job2.invoked)
	}
	if job3.invoked != 3 {
		t.Errorf("job3 invoked incorrect number of times: %d", job3.invoked)
	}
}

// At the moment, concur does nothing to avoid race conditions, so this
// test has been disabled. I don't believe it is possible for concur
// to avoid race conditions in any practical scenarios, so this will
// likely never be supported.
// func TestConcurrentRunnerRaceCondition(t *testing.T) {
// 	const numRuns = 100000
// 	mockJob := MockJob{}
// 	mockJobs := make([]*MockJob, numRuns)
// 	tasks := make([]concur.Task, len(mockJobs))
// 	for i := range mockJobs {
// 		// We want all of the jobs to be the same
// 		// That's how we'll test for a race condition
// 		mockJobs[i] = &mockJob
// 		tasks[i] = mockJobs[i]
// 	}

// 	err := concur.Concurrent().Run(tasks...)
// if err != nil {
// 	t.Error(err)
// }

// 	if mockJob.invokedUnatomic != numRuns {
// 		t.Errorf("mockJob.invoked incorrect number of times: %d", mockJob.invokedUnatomic)
// 	}
// }

func TestConcurrentRunnerReturnData(t *testing.T) {
	mockJobs := make([]MockJob, 42)
	tasks := make([]concur.Task, len(mockJobs))
	for i := range mockJobs {
		mockJobs[i].returnValue = fmt.Sprintf("MockJobReturnValue%d", i)
		tasks[i] = &mockJobs[i]
	}

	err := concur.Concurrent().Run(tasks...)
	if err != nil {
		t.Error(err)
	}

	for i, job := range mockJobs {
		if job.invoked != 1 {
			t.Errorf("jobs[%d] invoked incorrect number of times: %d", i, job.invoked)
		}
		if job.actualReturnValue != fmt.Sprintf("MockJobReturnValue%d", i) {
			t.Errorf("jobs[%d] returned incorrect results: %s", i, job.actualReturnValue)
		}
	}
}

func TestConcurrentRunnerPanic(t *testing.T) {
	const numRuns = 100
	mockJob := MockJob{}
	panicJob := PanicJob{}
	mockJobs := make([]*MockJob, numRuns)
	tasks := make([]concur.Task, len(mockJobs))
	for i := range mockJobs {
		// Halfway through, we'll have it panic
		if i == numRuns/2 {
			tasks[i] = &panicJob
		} else {
			mockJobs[i] = &mockJob
			tasks[i] = mockJobs[i]
		}
	}

	err := concur.Concurrent().Run(tasks...)
	if err == nil || !strings.Contains(err.Error(), "panic") {
		t.Errorf("received incorrect error: %v", err)
	}

	if mockJob.invoked != numRuns-1 {
		t.Errorf("mockJob.invoked incorrect number of times: %d", mockJob.invoked)
	}
}
