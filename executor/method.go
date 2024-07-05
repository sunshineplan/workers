package executor

import (
	"context"
	"errors"
	"math/rand/v2"

	"github.com/sunshineplan/workers"
)

// Method represents the execute method.
type Method int

const (
	// Concurrent method. Executing at the same time.
	Concurrent Method = iota
	// Serial method. Executing in order.
	Serial
	// Random method. Executing randomly.
	Random
)

// Executor is a generic struct for managing job execution.
type Executor[Arg, Res any] struct {
	w *workers.Workers
}

// New creates a new Executor with a specified number of workers.
// If n <= 0, it uses the default workers.
func New[Arg, Res any](n int64) *Executor[Arg, Res] {
	if n <= 0 {
		return &Executor[Arg, Res]{workers.DefaultWorkers}
	}
	return &Executor[Arg, Res]{workers.NewWorkers(n)}
}

// Execute gets the result from the functions with several args by specified method.
// If both argMethod and fnMethod is Concurrent, fnMethod will be first.
func (e *Executor[Arg, Res]) Execute(argMethod, fnMethod Method, args []Arg, fn ...func(Arg) (Res, error)) (res Res, err error) {
	if len(fn) == 0 {
		err = errors.New("no function provided")
		return
	}

	var clone []Arg
	var count int
	var nilArg bool
	switch len(args) {
	case 0:
		nilArg = true
	default:
		count = len(args)
		clone = make([]Arg, count)
		copy(clone, args)
	}

	switch fnMethod {
	case Concurrent:
		result := make(chan Res, 1)
		lasterr := make(chan error, 1)

		if nilArg {
			ctx := argContext[Arg, Res](len(fn), *new(Arg))
			defer ctx.cancel()

			e.w.Run(context.Background(), workers.SliceJob(fn, func(_ int, fn func(Arg) (Res, error)) {
				ctx.runFn(fn, result, lasterr)
			}))

			if err = <-lasterr; err != nil {
				return
			}

			return <-result, nil
		}

		if argMethod == Random {
			rand.Shuffle(count, func(i, j int) { clone[i], clone[j] = clone[j], clone[i] })
		}

		for i := range count {
			ctx := argContext[Arg, Res](len(fn), clone[i])
			e.w.Run(context.Background(), workers.SliceJob(fn, func(_ int, fn func(Arg) (Res, error)) {
				ctx.runFn(fn, result, lasterr)
			}))

			if err = <-lasterr; err == nil {
				res = <-result
				return
			} else if i == count-1 {
				return
			}
			ctx.cancel()
		}
	case Serial, Random:
		if fnMethod == Random {
			rand.Shuffle(len(fn), func(i, j int) { fn[i], fn[j] = fn[j], fn[i] })
		}

		for i, f := range fn {
			result := make(chan Res, 1)
			lasterr := make(chan error, 1)

			ctx := fnContext(count, f)
			if nilArg {
				ctx.runArg(*new(Arg), result, lasterr)
			} else {
				var worker *workers.Workers
				switch argMethod {
				case Concurrent:
					worker = e.w
				case Serial, Random:
					worker = workers.NewWorkers(1)
				default:
					err = errors.New("unknown arg method")
					return
				}

				if argMethod == Random {
					rand.Shuffle(count, func(i, j int) { clone[i], clone[j] = clone[j], clone[i] })
				}

				worker.Run(context.Background(), workers.SliceJob(clone, func(_ int, arg Arg) {
					ctx.runArg(arg, result, lasterr)
				}))
			}

			if err = <-lasterr; err == nil {
				res = <-result
				return
			} else if i == len(fn)-1 {
				return
			}
			ctx.cancel()
		}
	default:
		err = errors.New("unknown function method")
		return
	}
	err = errors.New("unknown error")
	return
}

// ExecuteConcurrentArg gets the fastest result from the functions with args, args will be run concurrently.
func (e *Executor[Arg, Res]) ExecuteConcurrentArg(arg []Arg, fn ...func(Arg) (Res, error)) (Res, error) {
	return e.Execute(Concurrent, Serial, arg, fn...)
}

// ExecuteConcurrentFunc gets the fastest result from the functions with args, functions will be run concurrently.
func (e *Executor[Arg, Res]) ExecuteConcurrentFunc(arg []Arg, fn ...func(Arg) (Res, error)) (Res, error) {
	return e.Execute(Serial, Concurrent, arg, fn...)
}

// ExecuteSerial gets the result until success from the functions with args in order.
func (e *Executor[Arg, Res]) ExecuteSerial(arg []Arg, fn ...func(Arg) (Res, error)) (Res, error) {
	return e.Execute(Serial, Serial, arg, fn...)
}

// ExecuteRandom gets the result until success from the functions with args randomly.
func (e *Executor[Arg, Res]) ExecuteRandom(arg []Arg, fn ...func(Arg) (Res, error)) (Res, error) {
	return e.Execute(Random, Random, arg, fn...)
}
