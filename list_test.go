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
			time.Sleep(750 * time.Millisecond)
			return
		}
		defer wg.Done()
		mu.Lock()
		log.Print(i)
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
	log.Print(res, res[:3], res[3:6], res[6:8])
	slices.Sort(res[:3])
	slices.Sort(res[3:6])
	slices.Sort(res[6:8])
	if expect := []int{0, 1, 2, 0, 4, 5, 7, 8}; !reflect.DeepEqual(expect, res) {
		t.Errorf("expected %v; got %v", expect, res)
	}
}
