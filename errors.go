package concur

type CumulativeError struct {
	Errors []error
}

func (c *CumulativeError) Add(err error) {
	c.Errors = append(c.Errors, err)
}
