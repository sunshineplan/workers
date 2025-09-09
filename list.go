package workers

import (
	"context"
	"errors"
	"sync"

	"github.com/sunshineplan/utils/container"
)

// JobList is a struct that holds a list of jobs, a worker pool, a function to execute jobs,
// a channel for signaling, and a boolean indicating if the job list is closed.
type JobList[T any] struct {
	mu     sync.Mutex
	l      *container.List[T]
	w      Workers
	f      func(T)
	c      chan struct{}
	closed bool
}

// NewJobList creates a new JobList with the given worker pool and job function.
func NewJobList[T any](workers int, f func(T)) *JobList[T] {
	return &JobList[T]{l: container.NewList[T](), w: Workers(workers), f: f}
}

// Start begins processing jobs in the job list using the provided context.
func (l *JobList[T]) Start(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return errors.New("job list is closed")
	}
	if l.c != nil {
		return errors.New("job list is already started")
	}
	l.c = make(chan struct{}, 1)
	c := make(chan func())
	l.w.Listen(ctx, c)
	go func() {
		defer close(c)
		for {
			select {
			case <-ctx.Done():
				l.Close()
				return
			case _, ok := <-l.c:
				if !ok {
					return
				}
				for {
					e := l.l.Front()
					if e == nil {
						break
					}
					v := l.l.Remove(e)
					c <- func() { l.f(v) }
				}
			}
		}
	}()
	return nil
}

// Close stops processing jobs and clears the job list.
func (l *JobList[T]) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return errors.New("job list is already closed")
	}
	l.l.Init()
	close(l.c)
	l.closed = true
	return nil
}

// PushBack adds a job to the end of the job list and signals the worker pool.
func (l *JobList[T]) PushBack(v T) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return errors.New("job list is closed")
	}
	if l.c == nil {
		return errors.New("job list is not started")
	}
	l.l.PushBack(v)
	select {
	case l.c <- struct{}{}:
	default:
	}
	return nil
}

// PushFront adds a job to the front of the job list and signals the worker pool.
func (l *JobList[T]) PushFront(v T) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return errors.New("job list is closed")
	}
	if l.c == nil {
		return errors.New("job list is not started")
	}
	l.l.PushFront(v)
	select {
	case l.c <- struct{}{}:
	default:
	}
	return nil
}
