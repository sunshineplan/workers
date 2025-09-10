package workers

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestFunction(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var n atomic.Int64
	if err := DefaultWorkers.Run(
		ctx,
		FuncJob(func() {
			if ctx.Err() == nil {
				n.Add(1)
				time.Sleep(2 * time.Second)
			}
		}),
	); err != nil && err != context.DeadlineExceeded {
		t.Fatal(err)
	}
	if expect, n := DefaultWorkers.weight(), n.Load(); n != expect {
		t.Errorf("expected %v; got %v", expect, n)
	}
}

func TestSlice(t *testing.T) {
	result := make([]string, 3)
	type test struct {
		char  string
		times int
	}
	if err := DefaultWorkers.Run(
		context.Background(),
		SliceJob(
			[]test{{"a", 1}, {"b", 2}, {"c", 3}},
			func(i int, item test) {
				result[i] = strings.Repeat(item.char, item.times)
			},
		),
	); err != nil {
		t.Fatal(err)
	}
	if expect := []string{"a", "bb", "ccc"}; !reflect.DeepEqual(expect, result) {
		t.Errorf("expected %v; got %v", expect, result)
	}
}

func TestMap(t *testing.T) {
	var m sync.Mutex
	var result []string
	if err := DefaultWorkers.Run(
		context.Background(),
		MapJob(
			map[string]int{"a": 1, "b": 2, "c": 3},
			func(k string, v int) {
				m.Lock()
				result = append(result, strings.Repeat(k, v))
				m.Unlock()
			},
		),
	); err != nil {
		t.Fatal(err)
	}
	sort.Strings(result)
	if expect := []string{"a", "bb", "ccc"}; !reflect.DeepEqual(expect, result) {
		t.Errorf("expected %v; got %v", expect, result)
	}
}

func TestRange(t *testing.T) {
	items := []string{"a", "b", "c"}
	expect := []string{"a", "bb", "ccc"}
	result := make([]string, 3)
	if err := DefaultWorkers.Run(
		context.Background(),
		RangeJob(
			1, 3,
			func(num int) {
				result[num-1] = strings.Repeat(items[num-1], num)
			},
		),
	); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("expected %v; got %v", expect, result)
	}
	result = make([]string, 3)
	if err := DefaultWorkers.Run(
		context.Background(),
		RangeJob(
			3, 1,
			func(num int) {
				result[num-1] = strings.Repeat(items[num-1], num)
			},
		),
	); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("expected %v; got %v", expect, result)
	}
}

func TestListen(t *testing.T) {
	var m sync.Mutex
	var result []string
	c := make(chan func(), 3)
	var wg sync.WaitGroup
	for k, v := range map[string]int{"a": 1, "b": 2, "c": 3} {
		wg.Add(1)
		c <- func() {
			defer wg.Done()
			m.Lock()
			result = append(result, strings.Repeat(k, v))
			m.Unlock()
		}
	}
	ctx := t.Context()
	DefaultWorkers.Listen(ctx, c)
	wg.Wait()
	sort.Strings(result)
	if expect := []string{"a", "bb", "ccc"}; !reflect.DeepEqual(expect, result) {
		t.Errorf("expected %v; got %v", expect, result)
	}
}
