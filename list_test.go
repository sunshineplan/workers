package workers

import (
	"context"
	"log"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestJobList(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	var wg sync.WaitGroup
	wg.Add(8)
	var mu sync.Mutex
	var res []int
	var list *JobList[int]
	test := func(i int) {
		if i != 0 && i%3 == 0 {
			list.PushFront(0)
			time.Sleep(300 * time.Millisecond)
			return
		}
		defer wg.Done()
		mu.Lock()
		res = append(res, i)
		mu.Unlock()
		time.Sleep(time.Second)
	}
	list = NewJobList(3, test)
	list.Start(context.Background())
	for i := range 8 {
		list.PushBack(i + 1)
	}
	wg.Wait()
	if n := list.l.Len(); n != 0 {
		t.Fatalf("expected 0; got %d", n)
	}
	log.Print(res, res[:2], res[2:5], res[5:8])
	slices.Sort(res[:2])
	slices.Sort(res[2:5])
	slices.Sort(res[5:8])
	if expect := []int{1, 2, 0, 4, 5, 0, 7, 8}; !reflect.DeepEqual(expect, res) {
		t.Errorf("expected %v; got %v", expect, res)
	}
}
