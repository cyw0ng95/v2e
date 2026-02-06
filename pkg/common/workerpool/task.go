package workerpool

import "context"

// Task represents a unit of work that can be executed by a worker
type Task interface {
	// Execute runs the task with the provided context
	// Returns an error if the task fails
	Execute(ctx context.Context) error
}

// TaskFunc is a function adapter that implements the Task interface
type TaskFunc func(ctx context.Context) error

// Execute implements the Task interface for TaskFunc
func (f TaskFunc) Execute(ctx context.Context) error {
	return f(ctx)
}
