package concur_test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/stevenmatthewt/concur"
)

type Job struct {
	data int
}

func (job *Job) Exec() error {
	time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
	job.data = rand.Int()
	return nil
}

// This is a simple example demonstrating how to make 3 calls in parallel.
// The calls in the example are simply returning random values, but
// in reality these could be external http calls, or any other moderately
// long running process.
func Example_simple() {
	job1 := &Job{}
	job2 := &Job{}
	job3 := &Job{}

	concur.Concurrent().Run(job1, job2, job3)

	fmt.Printf("job1 produced value: %d", job1.data)
	fmt.Printf("job1 produced value: %d", job1.data)
	fmt.Printf("job1 produced value: %d", job1.data)
}
