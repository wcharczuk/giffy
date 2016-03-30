Go-Web
======

[![Build Status](https://travis-ci.org/wcharczuk/go-web.svg?branch=master)](https://travis-ci.org/wcharczuk/go-web) [![GoDoc](https://godoc.org/github.com/wcharczuk/go-web?status.svg)](http://godoc.org/github.com/wcharczuk/go-web)

Go Web is a lightweight framework for building web applications in go. It rolls together very tightly scoped middleware with API endpoint and view endpoint patterns. 

##Example

Let's say we have a controller we need to implement:

```go
type FooController struct {}

func (fc FooController) barHandler(ctx *web.RequestContext) web.ControllerResult {
	return ctx.Raw([]byte("bar!"))
}

func (fc FooContoller) Register(app *web.App) {
	app.GET("/bar", fc.bar)
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
