package util

import (
	"reflect"
	"sync"
)

var (
	// Parallel contains parallel utils.
	Parallel = parallelUtil{}
)

type parallelUtil struct{}

func (pu parallelUtil) Each(collection interface{}, parallelism int, action func(interface{})) {
	t := reflect.TypeOf(collection)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(collection)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() != reflect.Slice {
		panic("cannot parallelize a non-slice.")
	}

	effectiveParallelism := parallelism
	if parallelism > v.Len() {
		effectiveParallelism = v.Len()
	}

	wg := sync.WaitGroup{}
	wg.Add(effectiveParallelism)

	for x := 0; x < v.Len(); x++ {
		action(v.Index(x).Interface())
	}
}

// Await waits for all actions to complete.
func (pu parallelUtil) Await(actions ...func()) {
	wg := sync.WaitGroup{}
	wg.Add(len(actions))
	for i := 0; i < len(actions); i++ {
		action := actions[i]
		go func() {
			action()
			wg.Done()
		}()
	}

	wg.Wait()
}
