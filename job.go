package workers

import "sync/atomic"

// Job interface defines a method to get the next job function.
type Job interface {
	Next() func()
}

// FuncJob type defines a function job.
type FuncJob func()

// Next returns the function itself for FuncJob.
func (job FuncJob) Next() func() {
	return job
}

// SliceJob creates a Job that iterates over a slice and applies a function to each element.
func SliceJob[T any](s []T, f func(int, T)) Job {
	return &sliceJob[T]{s: s, f: f}
}

// MapJob creates a Job that iterates over a map and applies a function to each key-value pair.
func MapJob[M ~map[K]V, K comparable, V any](m M, f func(K, V)) Job {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return &mapJob[M, K, V]{m: m, keys: r, f: f}
}

// Integer interface defines a set of integer types.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// RangeJob creates a Job that iterates over a range of integers and applies a function to each value.
func RangeJob[T Integer](start, end T, f func(T)) Job {
	n := new(atomic.Int64)
	n.Store(int64(start))
	return &rangeJob[T]{start: start, end: end, n: n, f: f}
}

var (
	_ Job = FuncJob(nil)
	_ Job = new(sliceJob[any])
	_ Job = new(mapJob[map[string]any, string, any])
	_ Job = new(rangeJob[int])
)

type sliceJob[T any] struct {
	s     []T
	index atomic.Int64
	f     func(int, T)
}

func (job *sliceJob[T]) Next() func() {
	n := int(job.index.Add(1)) - 1
	if n > len(job.s)-1 {
		return nil
	}
	return func() { job.f(n, job.s[n]) }
}

type mapJob[M ~map[K]V, K comparable, V any] struct {
	m     M
	keys  []K
	index atomic.Int64
	f     func(K, V)
}

func (job *mapJob[M, K, V]) Next() func() {
	n := int(job.index.Add(1)) - 1
	if n > len(job.keys)-1 {
		return nil
	}
	k := job.keys[n]
	return func() { job.f(k, job.m[k]) }
}

type rangeJob[T Integer] struct {
	start T
	end   T
	n     *atomic.Int64
	f     func(T)
}

func (job *rangeJob[T]) Next() func() {
	if job.start < job.end {
		if n := T(job.n.Add(1)) - 1; n <= job.end {
			return func() { job.f(n) }
		}
		return nil
	}
	if n := T(job.n.Add(-1)) + 1; n >= job.end {
		return func() { job.f(n) }
	}
	return nil
}
