Go-Web
======

[![Build Status](https://travis-ci.org/blendlabs/go-web.svg?branch=master)](https://travis-ci.org/blendlabs/go-web) [![GoDoc](https://godoc.org/github.com/blendlabs/go-web?status.svg)](http://godoc.org/github.com/blendlabs/go-web)

Go Web is a lightweight framework for building web applications in go. It rolls together very tightly scoped middleware with API endpoint and view endpoint patterns. 

##Requirements

* go 1.8+

##Example

Let's say we have a controller we need to implement:

```go
type FooController struct {}

func (fc FooContoller) Register(app *web.App) {
	app.GET("/bar", fc.bar)
}

func (fc FooController) barHandler(ctx *web.Ctx) web.Result {
	return ctx.Text().Result("bar!")
}
```

Then we would have the following in our `main.go`:

```go
func main() {
	app := web.New()
	app.Register(new(FooController))
	app.Start()
}
```

And that's it! There are options to configure things like the port and tls certificates, but the core use case is to bind
on 8080 or whatever is specified in the `PORT` environment variable. 

##Middleware

If you want to run some steps before controller actions fire (such as for auth etc.) you can add those steps as "middleware". 

```go
	app.GET("/admin/dashboard", c.dashboardAction, middle2, middle1, web.ViewProviderAsDefault)
```

This will then run `web.ViewProviderAsDefault` (which does something useful) and then call `middle1` and then `middle2`, finally the controller action.
An important detail is that the "cascading" nature of the calls depends on how you structure your middleware functions. If any of the middleware functions
return without calling the `action` parameter, execution stops there and subsequent middleware steps do not get called (ditto the controller action).

What do `middle1` and `middle2` look like? They are `ControllerMiddleware` and are just functions that take an action and return an action.

```go
func middle1(action web.ControllerAction) web.ControllerAction {
	return func(r *web.Ctx) web.ControllerResult {
		if r.Param("foo") != "bar" { //maximum security
			return r.DefaultResultProvider().NotAuthorized() //.DefaultResultProvider() is set by `web.ViewProviderAsDefault()`
															 // but also means we could use API or Text or just JSON endpoints with this middleware.
		}
		return action(r) //note we call the input action here!
	}
}
```

##Serving Static Files

You can set a path root to serve static files.

```go
func main() {
	app := web.New()
	app.Static("/static/*filepath", http.Dir("_client/dist"))
	app.Start()
}
```

A couple key points: we use the special token `*filepath` to denote a wildcard in the route that will be passed to the static file server (users of `httprouter` should recognize this, it's the same thing under the hood).
We then use the `http.Dir` function to specify the filesystem root that will be served (in this case a relative path).

You can also have a controller action return a static file:

```go
	app.GET("/thing", func(r *web.Ctx) web.ControllerResult { return r.Static("path/to/my/file") })
```

You can optionally set a static re-write rule (such as if you are cache-breaking assets with timestamps in the filename):

```go
func main() {
	app := web.New()
	app.Static("/static/*filepath", http.Dir("_client/dist"))
	app.StaticRewrite("/static/*filepath", `^(.*)\.([0-9]+)\.(css|js)$`, func(path string, parts ...string) string {
		return fmt.Sprintf("%s.%s", parts[1], parts[3])
	})
	app.Start()
}
```

Here we feed the `StaticRewrite` function the path (the same path as our static file server, this is important), a regular expression to match, and a special handler function that returns an updated path. 

Note: `parts ...string` is the regular expression sub matches from the expression, with `parts[0]` equal to the full input string. `parts[1]` and `parts[3]` in this case are the nominal root stem, and the extension respecitvely.

You can also set custom headers for static files:

```go
func main() {
	app := web.New()
	app.Static("/static/*filepath", http.Dir("_client/dist"))
	app.StaticRewrite("/static/*filepath", `^(.*)\.([0-9]+)\.(css|js)$`, func(path string, parts ...string) string {
		return fmt.Sprintf("%s.%s", parts[1], parts[3])
	})
	app.StaticHeader("/static/*filepath", "cache-control", "public,max-age=99999999")	
}
```

This will then set the specified cache headers on response for the static files. 

##Benchmarks

Benchmarks are key, obviously, because the ~200us you save choosing a framework won't be wiped out by the 50ms ping time to your servers. 

For a relatively clean implementation (found in `benchmark/main.go`) that uses `go-web`:
```
Running 10s test @ http://localhost:9090/json
  2 threads and 64 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     0.92ms  223.27us  11.82ms   86.00%
    Req/Sec    34.19k     2.18k   40.64k    65.35%
  687017 requests in 10.10s, 203.11MB read
Requests/sec:  68011.73
Transfer/sec:     20.11MB
```

On the same machine, with a very, very bare bones implementation using only built-in stuff in `net/http`:

```
Running 10s test @ http://localhost:9090/json
  2 threads and 64 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     0.93ms  216.64us  10.25ms   89.10%
    Req/Sec    34.22k     2.73k   40.63k    59.90%
  687769 requests in 10.10s, 109.54MB read
Requests/sec:  68091.37
Transfer/sec:     10.84MB
```

The key here is to make sure not to enable logging, because if logging is enabled that throughput gets cut in half. 