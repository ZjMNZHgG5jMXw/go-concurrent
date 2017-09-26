package concurrent

import (
	"reflect"
	"sync"
)

// MakeMerge (&func(...chan int) chan int)
func MakeMerge(in interface{}) {
	inT := reflect.TypeOf(in).Elem()
	fstT := inT.In(0).Elem()
	dstT := reflect.FuncOf(
		[]reflect.Type{reflect.SliceOf(fstT)},
		[]reflect.Type{fstT},
		true)
	dstV := reflect.MakeFunc(
		dstT,
		func(args []reflect.Value) (res []reflect.Value) {
			var (
				resV      = reflect.MakeChan(reflect.ChanOf(reflect.BothDir, fstT.Elem()), 0)
				chsV      = args[0]
				allClosed sync.WaitGroup
				done      = make(chan struct{})
				cases     = make([]reflect.SelectCase, chsV.Len()+1)
			)

			for i := 0; i < chsV.Len(); i++ {
				cases[i] = reflect.SelectCase{
					Dir:  reflect.SelectRecv,
					Chan: chsV.Index(i)}
			}
			cases[chsV.Len()] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(done)}

			allClosed.Add(chsV.Len())

			go func() {
				defer close(done)
				allClosed.Wait()
				done <- struct{}{}
			}()

			go func() {
				defer resV.Close()
				var (
					value  reflect.Value
					ok     bool
					i      int
					closed = make(map[int]struct{})
				)

				for {
					i, value, ok = reflect.Select(cases)
					if ok {
						if i == chsV.Len() {
							return
						}
						resV.Send(value)
					} else {
						if _, ok := closed[i]; !ok {
							closed[i] = struct{}{}
							allClosed.Done()
						}
					}
				}
			}()

			return []reflect.Value{resV}
		})

	reflect.ValueOf(in).Elem().Set(dstV)
}
