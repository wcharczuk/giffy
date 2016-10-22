package workQueue

import "fmt"

// Entry is an individual item of work.
type Entry struct {
	Action Action
	Args   []interface{}
	Tries  int
}

func (e Entry) String() string {
	return fmt.Sprintf("{ %#v args: %v tries: %d }", e.Action, e.Args, e.Tries)
}
