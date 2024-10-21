package workers

import (
	"context"
	"runtime"

	"golang.org/x/sync/semaphore"
)

// DefaultWorkers is a default instance of Workers with the size of GOMAXPROCS.
var DefaultWorkers = Workers(runtime.GOMAXPROCS(0))

// Workers holds the size and semaphore for managing concurrent jobs.
type Workers int

func (i Workers) weight() int64 {
	if i <= 0 {
		return int64(runtime.GOMAXPROCS(0))
	}
	return int64(i)
}

// Run executes jobs from the Job interface until there are no more jobs.
// It acquires a semaphore weight for each job and releases it when the job is done.
func (i Workers) Run(ctx context.Context, job Job) (err error) {
	weight := i.weight()
	w := semaphore.NewWeighted(weight)
	for f := job.Next(); f != nil; f = job.Next() {
		if err = w.Acquire(ctx, 1); err != nil {
			return
		}
		go func() {
			defer w.Release(1)
			f()
		}()
	}
	if err = w.Acquire(ctx, weight); err == nil {
		w.Release(weight)
	}
	return
}

// Listen listens for jobs from a channel and runs them concurrently.
// It stops listening when the context is done or the channel is closed.
func (i Workers) Listen(ctx context.Context, c <-chan func()) {
	w := semaphore.NewWeighted(i.weight())
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
				if err := w.Acquire(ctx, 1); err != nil {
					return
				}
				go func() {
					defer w.Release(1)
					job()
				}()
			}
		}
	}()
}

// Run executes jobs using DefaultWorkers from the Job interface until there are no more jobs.
// It acquires a semaphore weight for each job and releases it when the job is done.
func Run(ctx context.Context, job Job) error {
	return DefaultWorkers.Run(ctx, job)
}

// Listen listens for jobs using DefaultWorkers from a channel and runs them concurrently.
// It stops listening when the context is done or the channel is closed.
func Listen(ctx context.Context, c <-chan func()) {
	DefaultWorkers.Listen(ctx, c)
}
