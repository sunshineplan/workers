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
	var res []int
	var mu sync.Mutex
	c := make(chan struct{}, 3)
	test := func(i int) {
		time.Sleep(time.Second)
		if i != 0 && i%3 == 0 {
			c <- struct{}{}
			return
		}
		log.Print(i)
		mu.Lock()
		defer mu.Unlock()
		res = append(res, i)
		wg.Done()
	}
	list := NewJobList(NewWorkers(3), test)
	list.Start(context.Background())
	go func() {
		for {
			<-c
			list.PushFront(0)
		}
	}()
	for i := range 8 {
		list.PushBack(i + 1)
	}
	wg.Wait()
	if n := list.l.Len(); n != 0 {
		t.Errorf("expected 0; got %d", n)
	}
	slices.Sort(res[:2])
	slices.Sort(res[2:5])
	slices.Sort(res[5:7])
	if expect := []int{1, 2, 0, 4, 5, 7, 8, 0}; !reflect.DeepEqual(expect, res) {
		t.Errorf("expected %v; got %v", expect, res)
	}
}
