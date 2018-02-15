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

// ConcurrentRunner runs Tasks concurrently.
type ConcurrentRunner struct {
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
	timer := time.NewTimer(time.Second * 10)

	runner.errorChan = make(chan error)
	runner.successChan = make(chan bool)
	runner.timeoutChan = timer.C
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
	timer.Stop()

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
			break
		}
	}
	if cumulativeErr.isError() {
		return cumulativeErr
	}
	return nil
}
