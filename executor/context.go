package executor

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrSkip       = errors.New("skip")
	ErrAllSkipped = errors.New("all skipped")
)

type key int

const (
	fnKey key = iota + 1
	argKey
)

type fn[Arg, Res any] interface {
	func(Arg) (Res, error)
}

type Context[Arg, Res any] struct {
	context.Context
	cancel context.CancelFunc

	mu  sync.Mutex
	res []error

	count int
}

func newContext[Arg, Res any](ctx context.Context, count int) *Context[Arg, Res] {
	ctx, cancel := context.WithCancel(ctx)
	return &Context[Arg, Res]{ctx, cancel, sync.Mutex{}, nil, count}
}

func fnContext[Arg, Res any, Fn fn[Arg, Res]](count int, fn Fn) *Context[Arg, Res] {
	return newContext[Arg, Res](context.WithValue(context.Background(), fnKey, fn), count)
}

func argContext[Arg, Res any](count int, arg Arg) *Context[Arg, Res] {
	return newContext[Arg, Res](context.WithValue(context.Background(), argKey, arg), count)
}

func (ctx *Context[Arg, Res]) run(executor func(chan<- Res, chan<- error), rc chan<- Res, ec chan<- error) {
	if ctx.Err() != nil {
		return
	}

	r := make(chan Res, 1)
	c := make(chan error, 1)
	go executor(r, c)

	select {
	case <-ctx.Done():
		return
	case err := <-c:
		ctx.mu.Lock()
		defer ctx.mu.Unlock()

		if err != nil {
			if err != ErrSkip {
				ctx.res = append(ctx.res, err)
			}

			if ctx.count <= 1 {
				rc <- *new(Res)
				if l := len(ctx.res); l > 0 {
					ec <- ctx.res[l-1]
				} else {
					ec <- ErrAllSkipped
				}
			}
			ctx.count--
		} else {
			ctx.cancel()

			rc <- <-r
			ec <- nil
		}
	}
}

func (ctx *Context[Arg, Res]) runArg(arg Arg, rc chan<- Res, ec chan<- error) {
	ctx.run(func(c1 chan<- Res, c2 chan<- error) {
		r, err := (ctx.Value(fnKey).(func(Arg) (Res, error)))(arg)
		c1 <- r
		c2 <- err
	}, rc, ec)
}

func (ctx *Context[Arg, Res]) runFn(fn func(Arg) (Res, error), rc chan<- Res, ec chan<- error) {
	ctx.run(func(c1 chan<- Res, c2 chan<- error) {
		r, err := fn(ctx.Value(argKey).(Arg))
		c1 <- r
		c2 <- err
	}, rc, ec)
}
