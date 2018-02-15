package concur

// Task is an interface describing enything that
// can be executed by the concur package.
// It has only one method requirement: Exec()
type Task interface {
	// Exec is the method that will be called when the task is executed
	Exec() error
}
