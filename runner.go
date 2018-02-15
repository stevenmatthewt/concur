package concur

import (
	"fmt"
	"time"

	"errors"
)

// Runner is an interface describing anything
// that is capable of running Tasks.
type Runner interface {
	Run(tasks ...Task) error
}

// Concurrent creates a new ConcurrentRunner
// which is responsible for running Tasks
// concurrently.
func Concurrent() ConcurrentRunner {
	return ConcurrentRunner{}
}

// SetTimeout sets the timeout for the current Runner
// If execution of any Tasks exceeds the timeout, the Runner will return an error.
// Please not that the remaining Tasks will continue to execute in their own goroutine
func (runner ConcurrentRunner) SetTimeout(dur time.Duration) ConcurrentRunner {
	runner.timeout = &dur
	return runner
}

// ConcurrentRunner runs Tasks concurrently.
type ConcurrentRunner struct {
	timeout     *time.Duration
	successChan chan bool
	errorChan   chan error
	timeoutChan <-chan time.Time
}

// Run takes a list of Tasks and runs them concurrently.
// An error is returned if any Tasks return an error.
// Please note that one Task returning an error
// will not halt execution of the remaining Tasks.
//
// Run block until execution of all Tasks is complete.
func (runner ConcurrentRunner) Run(tasks ...Task) (err error) {
	if runner.timeout != nil {
		timer := time.NewTimer(*runner.timeout)
		runner.timeoutChan = timer.C
		defer timer.Stop()
	}

	runner.errorChan = make(chan error)
	runner.successChan = make(chan bool)
	for _, task := range tasks {
		go func(task Task) {
			defer func() {
				if r := recover(); r != nil {
					switch x := r.(type) {
					case string:
						err = errors.New(x)
					case error:
						err = x
					default:
						err = fmt.Errorf("%+v", r)
					}
					runner.errorChan <- err
				}
			}()
			err := task.Exec()
			if err != nil {
				runner.errorChan <- err
			} else {
				runner.successChan <- true
			}
		}(task)
	}

	err = runner.waitOnChannels(len(tasks))

	return err
}

func (runner ConcurrentRunner) waitOnChannels(num int) error {
	var cumulativeErr CumulativeError
	for i := 0; i < num; {
		select {
		case err := <-runner.errorChan:
			cumulativeErr.add(err)
			i++
		case <-runner.successChan:
			i++
		case <-runner.timeoutChan:
			cumulativeErr.add(errors.New("timed out waiting for task(s) to complete"))
			close(runner.errorChan)
			close(runner.successChan)
			return cumulativeErr
		}
	}
	if cumulativeErr.isError() {
		return cumulativeErr
	}
	return nil
}
