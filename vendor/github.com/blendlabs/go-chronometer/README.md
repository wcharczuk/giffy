Chronometer
===========

[![Build Status](https://travis-ci.org/blendlabs/go-chronometer.svg?branch=master)](https://travis-ci.org/blendlabs/go-chronometer)

Chronometer is a basic job scheduling, task handling library that wraps goroutines with a little metadata.

##Getting Started

Here is a simple example of getting started with chronometer

```go
package main

import "github.com/blendlabs/go-chronometer"

...

	chronometer.Default().RunTask(chronometer.NewTask(func(ct *chronometer.CancellationToken) error {
		... //long winded process here
		ct.CheckCancellation()
	}))
```

A couple things going on here. First, we're accessing a centralized `JobManager` instance through `Default()`. We're then running a task, and creating a quick task that is simply a wrapper on a function signature. The `*CancellationToken` is the mechanism the manager uses to tell the task to abort. 

For a more detailed (running) example, look in `sample/main.go`.

###What is a `CancellationToken`

It is the mechanism by which we signal that a task should abort. We don't have a reference to a thread or similar, so we use an object and signal with a boolean. We could use channels, but this is simpler. 

###Schedules

Schedules are very basic right now, either the job runs on a fixed interval (every minute, every 2 hours etc) or on given days weekly (every day at a time, or once a week at a time).

You're free to implement your own schedules outside the basic ones; a schedule is just an interface for `GetNextRunTime(after time.Time)`.

###Tasks vs. Jobs

Jobs are tasks with schedules, thats about it. The interfaces are very similar otherwise. 
