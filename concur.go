package concur

type Task interface {
	Exec() error
}
