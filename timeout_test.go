package concur_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stevenmatthewt/concur"
)

func TestConcurrentRunnerTimeout(t *testing.T) {
	jobs := []MockJob{
		MockJob{
			sleepDuration: time.Millisecond * 10000,
		},
		MockJob{},
	}

	err := concur.Concurrent().SetTimeout(time.Millisecond*5).Run(&jobs[0], &jobs[1])
	if err == nil {
		t.Fatal("Expected error but did not receive one")
	}
	if !strings.Contains(err.Error(), "time") {
		t.Errorf("Error message in wrong format: %v", err)
	}
}
