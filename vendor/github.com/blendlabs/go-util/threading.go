package util

import "sync"

// AwaitAction is a function that can be "awaited".
type AwaitAction func()

// AwaitAll waits for all actions to complete.
func AwaitAll(actions ...AwaitAction) {
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
