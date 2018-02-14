package concur

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type Runner interface {
	Run(tasks ...Task) error
}

func Concurrent() ConcurrentRunner {
	return ConcurrentRunner{}
}

type ConcurrentRunner struct {
	successChan chan bool
	errorChan   chan error
	timeoutChan <-chan time.Time
}

// Run takes a list of tasks and runs them concurrently.
// An error is returned if any tasks return an error.
// Please note that one task returning an error
// will not halt execution of the remaining tasks
func (runner ConcurrentRunner) Run(tasks ...Task) (err error) {
	timer := time.NewTimer(time.Second)

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
					runner.errorChan <- errors.Wrap(err, "task panicked")
				}
			}()
			err := task.Exec()
			if err != nil {
				runner.errorChan <- errors.Wrap(err, "failed to run task")
			}
			runner.successChan <- true
		}(task)
	}

	err = runner.waitOnChannels(len(tasks))
	timer.Stop()

	return err
}

func (runner ConcurrentRunner) waitOnChannels(num int) error {
	var err error
	for i := 0; i < num; {
		select {
		case err = <-runner.errorChan:
			i++
		case <-runner.successChan:
			i++
		case <-runner.timeoutChan:
			err = errors.New("timed out waiting for task(s) to complete")
			break
		}
	}
	return err
}
