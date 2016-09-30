Go Logger
=========

Go logger is not well named. It's really an event queue that is managed by a bitmask flag set.

# Example

```golang
logger.SetDiagnostics(logger.NewDiagnosticsAgentFromEnvironment()) // set the singleton to a environment configured default.
logger.Diagnostics().EventQueue().UseSynchronousDispatch() //events fire in order, but will hang if queue is at capacity.
logger.Diagnostics().EventQueue().SetMaxWorkItems(1 << 20) //make the queue size enormous (~1mm items).
logger.Diagnostics().AddEventListener(logger.EventError, func(wr logger.Logger, ts logger.TimeSource, e uint64, args ...interface{}) {
    //ping an external service?
    //log something to the db?
    //this action will be handled by a separate go-routine
})
```

Then, elsewhere in our code:

```golang
logger.Diagnostics().Error(exception.New("this is an exception"))   // this will write the error to stderr, but also
                                                                    // will trigger the handler from before.
```

# What can I do with this?

You can defer writing a bunch of log messages to stdout to unblock requests in high-throughput scenarios. `logger` is very
careful to preserve timing state so that actions that live in the queue for multiple seconds are logged with the correct 
timestamp.

# What else can I do with this?

You can standardize how you write log messages across multiple packages / services.