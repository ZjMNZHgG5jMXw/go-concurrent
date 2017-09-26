package concurrent

import (
	"testing"
	"time"
)

func TestMakeMerge(t *testing.T) {
	var merge func(...chan int) chan int
	MakeMerge(&merge)

	var (
		ch1     = make(chan int)
		ch2     = make(chan int)
		occured = make(map[int]struct{})
	)
	ch3 := merge(ch1, ch2)

	go func() {
		defer close(ch1)
		ch1 <- 1
		time.Sleep(10 * time.Millisecond)
		ch1 <- 4
	}()

	go func() {
		defer close(ch2)
		ch2 <- 2
		time.Sleep(5 * time.Millisecond)
		ch2 <- 3
	}()

	for value := range ch3 {
		t.Log(value)
		if _, ok := occured[value]; ok {
			t.Errorf("found %d multiple times", value)
		}
		occured[value] = struct{}{}
	}

	for _, num := range []int{1, 2, 3, 4} {
		if _, ok := occured[num]; !ok {
			t.Errorf("%d did not occur", num)
		}
	}
}
