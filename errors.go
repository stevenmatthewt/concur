package concur

import (
	"fmt"
	"strings"
)

// CumulativeError is an error that records all
// individual errors that occur during task execution.
// These errors are then returned as a single CumulativeError
type CumulativeError struct {
	Errors []error
}

func (c *CumulativeError) add(err error) {
	c.Errors = append(c.Errors, err)
}

func (c *CumulativeError) isError() bool {
	return len(c.Errors) > 0
}

func (c CumulativeError) Error() string {
	var errstrings = make([]string, len(c.Errors))
	for i, err := range c.Errors {
		errstrings[i] = err.Error()
	}
	return fmt.Sprintf("errors occured during task execution: %s", strings.Join(errstrings, ", "))
}

// Expand returns a slice of all the errors included
// in the CumulativeError
func (c CumulativeError) Expand() []error {
	return c.Errors
}
