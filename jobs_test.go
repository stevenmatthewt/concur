package concur_test

import (
	"sync/atomic"
	"time"
)

type UnatomicMockJob struct {
	invoked int
}

func (job *UnatomicMockJob) Exec() error {
	job.invoked++
	return nil
}

type MockJob struct {
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
	atomic.AddInt32(&job.invoked, 1)
	// fmt.Println(job.returnError)
	if job.returnValue != "" {
		job.actualReturnValue = job.returnValue
	}
	return job.returnError
}

type PanicJob struct{}

func (job *PanicJob) Exec() error {
	panic("PanicJob panicking on purpose")
}
