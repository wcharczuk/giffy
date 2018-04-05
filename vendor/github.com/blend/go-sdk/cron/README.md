cron
====

`cron` is a basic job scheduling, task handling library that wraps goroutines with a little metadata.

## Getting Started

Here is a simple example of getting started with chronometer; running a background task with easy cancellation.

```go
package main

import "github.com/blend/go-sdk/cron"

...
	mgr := cron.New()
	mgr.RunTask(cron.NewTask(func(ctx context.Context) error {
		for {
			select {
			case <- ctx.Done():
				return nil
			default:
			... //long winded process here
				return nil
		}
	}))
```

The above wraps a simple function signature with a task, and allows us to check if a cancellation signal has been sent. 
For a more detailed (running) example, look in `_sample/main.go`.

### Schedules

Schedules are very basic right now, either the job runs on a fixed interval (every minute, every 2 hours etc) or on given days weekly (every day at a time, or once a week at a time).

You're free to implement your own schedules outside the basic ones; a schedule is just an interface for `GetNextRunTime(after time.Time)`.

### Tasks vs. Jobs

Jobs are tasks with schedules, thats about it. The interfaces are very similar otherwise. 

### Optional Interfaces

You can optionally implement interfaces to give you more control over your jobs:

```golang
Enabled() bool
```

or tasks:

```golang
SequentialExecution() bool
```

Allows you to enable or disable your job within the job itself; this allows all the code required to manage the job be in the same place.
