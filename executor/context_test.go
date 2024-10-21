package executor

import (
	"errors"
	"testing"
)

func TestSkip(t *testing.T) {
	tmp := errors.New("error")
	w := Executor[int, any](0)
	if _, err := w.ExecuteSerial(
		[]int{0, 1, 2},
		func(n int) (any, error) {
			if n == 0 {
				return nil, tmp
			}
			return nil, ErrSkip
		},
	); err != tmp {
		t.Errorf("expected %d; got %v", tmp, err)
	}

	if _, err := w.ExecuteSerial(
		[]int{0, 1, 2},
		func(n int) (any, error) {
			return nil, ErrSkip
		},
	); err != ErrAllSkipped {
		t.Errorf("expected %s; got %v", ErrAllSkipped, err)
	}
}
