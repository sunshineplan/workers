package workers

import (
	"context"
	"runtime"

	"golang.org/x/sync/semaphore"
)

// DefaultWorkers is a default instance of Workers with the size of GOMAXPROCS.
var DefaultWorkers = NewWorkers(int64(runtime.GOMAXPROCS(0)))

// Workers struct holds the size and semaphore for managing concurrent jobs.
type Workers struct {
	size int64
	w    *semaphore.Weighted
}

// NewWorkers creates a new Workers instance with a given size.
func NewWorkers(n int64) *Workers {
	return &Workers{n, semaphore.NewWeighted(n)}
}

// Size returns the size of the Workers.
func (w Workers) Size() int64 {
	return w.size
}

// Run executes jobs from the Job interface until there are no more jobs.
// It acquires a semaphore weight for each job and releases it when the job is done.
func (w *Workers) Run(ctx context.Context, job Job) (err error) {
	for f := job.Next(); f != nil; f = job.Next() {
		if err = w.w.Acquire(ctx, 1); err != nil {
			return
		}
		go func() {
			defer w.w.Release(1)
			f()
		}()
	}
	if err = w.w.Acquire(ctx, w.size); err == nil {
		w.w.Release(w.size)
	}
	return
}

// Listen listens for jobs from a channel and runs them concurrently.
// It stops listening when the context is done or the channel is closed.
func (w *Workers) Listen(ctx context.Context, c <-chan func()) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job, ok := <-c:
				if !ok {
					return
				}
				if job == nil {
					continue
				}
				if err := w.w.Acquire(ctx, 1); err != nil {
					return
				}
				go func() {
					defer w.w.Release(1)
					job()
				}()
			}
		}
	}()
}
