package executor

import (
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestExecuteConcurrent1(t *testing.T) {
	w := Executor[int, any](0)
	result, err := w.ExecuteConcurrentArg(
		[]int{0, 1, 2},
		func(n int) (any, error) {
			time.Sleep(time.Second * time.Duration(n))
			return n * 2, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if expect := 0; result != expect {
		t.Errorf("expected %d; got %v", expect, result)
	}

	_, err = w.ExecuteConcurrentArg(
		[]int{0, 1, 2},
		func(n int) (any, error) {
			time.Sleep(time.Second * time.Duration(n))
			return nil, fmt.Errorf("%v", n*2)
		},
	)
	if expect := "4"; err.Error() != expect {
		t.Errorf("expected error %s; got %v", expect, err)
	}
}

func TestExecuteConcurrent2(t *testing.T) {
	w := Executor[int, any](0)
	result, err := w.ExecuteConcurrentFunc(
		[]int{1},
		func(n int) (any, error) {
			return n * 0 * 2, nil
		},
		func(n int) (any, error) {
			time.Sleep(time.Second * 1 * time.Duration(n))
			return n * 1 * 2, nil
		},
		func(n int) (any, error) {
			time.Sleep(time.Second * 2 * time.Duration(n))
			return n * 2 * 2, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if expect := 0; result != expect {
		t.Errorf("expected %d; got %v", expect, result)
	}

	_, err = w.ExecuteConcurrentFunc(
		[]int{1},
		func(n int) (any, error) {
			return nil, fmt.Errorf("%v", n*0*2)
		},
		func(n int) (any, error) {
			time.Sleep(time.Second * 1 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*1*2)
		},
		func(n int) (any, error) {
			time.Sleep(time.Second * 2 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*2*2)
		},
	)
	if expect := "4"; err.Error() != expect {
		t.Errorf("expected error %s; got %v", expect, err)
	}
}

func TestExecuteSerial1(t *testing.T) {
	w := Executor[int, any](0)
	result, err := w.ExecuteSerial(
		[]int{0, 1, 2},
		func(n int) (any, error) {
			time.Sleep(time.Second * time.Duration(n))
			return n * 2, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if expect := 0; result != expect {
		t.Errorf("expected %d; got %v", expect, result)
	}

	_, err = w.ExecuteSerial(
		[]int{0, 1, 2},
		func(n int) (any, error) {
			time.Sleep(time.Second * time.Duration(n))
			return nil, fmt.Errorf("%v", n*2)
		},
	)
	if expect := "4"; err.Error() != expect {
		t.Errorf("expected error %s; got %v", expect, err)
	}
}

func TestExecuteSerial2(t *testing.T) {
	w := Executor[int, any](0)
	result, err := w.ExecuteSerial(
		[]int{1},
		func(n int) (any, error) {
			return n * 0 * 2, nil
		},
		func(n int) (any, error) {
			time.Sleep(time.Second * 1 * time.Duration(n))
			return n * 1 * 2, nil
		},
		func(n int) (any, error) {
			time.Sleep(time.Second * 2 * time.Duration(n))
			return n * 2 * 2, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if expect := 0; result != expect {
		t.Errorf("expected %d; got %v", expect, result)
	}

	_, err = w.ExecuteSerial(
		[]int{1},
		func(n int) (any, error) {
			return nil, fmt.Errorf("%v", n*0*2)
		},
		func(n int) (any, error) {
			time.Sleep(time.Second * 1 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*1*2)
		},
		func(n int) (any, error) {
			time.Sleep(time.Second * 2 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*2*2)
		},
	)
	if expect := "4"; err.Error() != expect {
		t.Errorf("expected error %s; got %v", expect, err)
	}
}

func TestExecuteRandom1(t *testing.T) {
	w := Executor[string, any](0)
	testcase := []string{"a", "b", "c", "d", "e", "f", "g"}
	var result []string
	if _, err := w.ExecuteRandom(
		testcase,
		func(i string) (any, error) {
			result = append(result, i)
			return nil, fmt.Errorf("%v", i)
		},
	); err == nil {
		t.Fatal("expected error; got nil")
	}
	if reflect.DeepEqual(testcase, result) {
		// test again
		result = nil
		if _, err := w.ExecuteRandom(
			testcase,
			func(i string) (any, error) {
				result = append(result, i)
				return nil, fmt.Errorf("%v", i)
			},
		); err == nil {
			t.Fatal("expected error; got nil")
		}
		if reflect.DeepEqual(testcase, result) {
			t.Error("expected not equal; got equal")
		}
	}
	slices.Sort(result)
	if !reflect.DeepEqual(testcase, result) {
		t.Errorf("expected %v; got %v", testcase, result)
	}
}

func TestExecuteRandom2(t *testing.T) {
	w := Executor[any, any](0)
	var result []string
	_, err := w.ExecuteRandom(
		nil,
		func(i any) (any, error) {
			result = append(result, "a")
			return nil, fmt.Errorf("a")
		},
		func(i any) (any, error) {
			result = append(result, "b")
			return nil, fmt.Errorf("b")
		},
		func(i any) (any, error) {
			result = append(result, "c")
			return nil, fmt.Errorf("c")
		},
		func(i any) (any, error) {
			result = append(result, "d")
			return nil, fmt.Errorf("d")
		},
		func(i any) (any, error) {
			result = append(result, "e")
			return nil, fmt.Errorf("e")
		},
		func(i any) (any, error) {
			result = append(result, "f")
			return nil, fmt.Errorf("f")
		},
		func(i any) (any, error) {
			result = append(result, "g")
			return nil, fmt.Errorf("g")
		},
	)
	if err == nil {
		t.Fatal("expected error; got nil")
	}

	expect := []string{"a", "b", "c", "d", "e", "f", "g"}
	if reflect.DeepEqual(expect, result) {
		t.Errorf("expected not equal; got equal: %v", result)
	}
	slices.Sort(result)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("expected equal; got not equal: %v", result)
	}
}

func TestLimit(t *testing.T) {
	w := Executor[int, any](0)
	_, err := w.Execute(
		Serial,
		Concurrent,
		[]int{1},
		func(n int) (any, error) {
			return nil, fmt.Errorf("%v", n*0*2)
		},
		func(n int) (any, error) {
			time.Sleep(time.Second * 1 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*1*2)
		},
		func(n int) (any, error) {
			time.Sleep(time.Second * 2 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*2*2)
		},
	)
	if expect := "4"; err.Error() != expect {
		t.Errorf("expected error %s; got %v", expect, err)
	}

	w = Executor[int, any](1)
	_, err = w.Execute(
		Serial,
		Concurrent,
		[]int{1},
		func(n int) (any, error) {
			return nil, fmt.Errorf("%v", n*0*2)
		},
		func(n int) (any, error) {
			time.Sleep(time.Second * 1 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*1*2)
		},
		func(n int) (any, error) {
			time.Sleep(time.Second * 2 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*2*2)
		},
	)
	if expect := "4"; err.Error() != expect {
		t.Errorf("expected error %s; got %v", expect, err)
	}
}
