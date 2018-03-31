package web

import (
	"bytes"
	"crypto/tls"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-logger"
	util "github.com/blendlabs/go-util"
	env "github.com/blendlabs/go-util/env"
)

// test keys
const (
	testCert = `-----BEGIN CERTIFICATE-----
MIIC+jCCAeKgAwIBAgIRAKGQgEUjhTZMM2VMx9y92MUwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAeFw0xODAzMDkwMzIxMzhaFw0xOTAzMDkwMzIx
MzhaMBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQDCONjExGZ+MwYZ1CosUB+sa9jS/AD0YkOi8AgiOYughLrKx5RuSsO9
ZaO0iwH987SFwAxBEiXwfLceEDgHYLGNfKQdYMCdh1yclr9yKrfpLV1SvPwT/utm
ek3ONwbJwqIrBP0dNWtfRhHhu2Gyc1JjxpqETdCUUZfuJWouVjVIxaIxLvyxYkUo
AS6SpUlUOOF3Wnre4+3x1RWRpXwns/HUFjsQBOIBo7pganxzcukTsQZWv+kJEA2o
EW33VdLQBuD59X6h1/qjx93s3AndeT5CoeVCAQ6PKXuV9z1WCpRewPpD+J89Noff
aueXIhTvxpFnB6W6VGVDQmnhEbnwA2IPAgMBAAGjSzBJMA4GA1UdDwEB/wQEAwIF
oDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMBQGA1UdEQQNMAuC
CWxvY2FsaG9zdDANBgkqhkiG9w0BAQsFAAOCAQEAYkkoNdditdKaEWrUjMc52QqJ
e4hbjqWT6W3bphGgYiKvnxgcDQYL3+RgEd7tGIHfgLkIiuM9efH+KJ4/jdXFWlcQ
7PoS9nGn0FwNvGdt9KCzNZSODSgQNt7FdsSpfw6Qzhn6XCwx3Bay9uF6cPap+wtX
SX6fD+az+dh0UPYoEltuKBv43+wLwsxAg18vBFuACI52NomvNw4G4uw4epBGGmp8
A0A4h9O67T/bFXchS+uIQnThZo4U/TCDu0xi/Q89xtjWff1YybwR85l85pEt1v7G
ei1eKWKYUxUU7lBMaECknLsJ4xsDKRSA5tvEDCkeQDCwTD7Msh5uGQ9itoWMlQ==
-----END CERTIFICATE-----
`
	testKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAwjjYxMRmfjMGGdQqLFAfrGvY0vwA9GJDovAIIjmLoIS6yseU
bkrDvWWjtIsB/fO0hcAMQRIl8Hy3HhA4B2CxjXykHWDAnYdcnJa/ciq36S1dUrz8
E/7rZnpNzjcGycKiKwT9HTVrX0YR4bthsnNSY8aahE3QlFGX7iVqLlY1SMWiMS78
sWJFKAEukqVJVDjhd1p63uPt8dUVkaV8J7Px1BY7EATiAaO6YGp8c3LpE7EGVr/p
CRANqBFt91XS0Abg+fV+odf6o8fd7NwJ3Xk+QqHlQgEOjyl7lfc9VgqUXsD6Q/if
PTaH32rnlyIU78aRZwelulRlQ0Jp4RG58ANiDwIDAQABAoIBAHcg8yTN6qfhmA5j
qnJ/us3BYL8Yv2UmmKHqZLLJZTFR+FjEzfBQf3s+SolE8jXYM5QOVfXbsdWuSYtx
G0y7LGzCVM+INtzo2A9cD5VxSlkF8EX9kQiaxbyXq/2eltVOQrXsW2x9BZzsl69D
hgs03QZCHSilqhgva+cwn85IJmq5bL5BMlNT1vFUgKz4QWISuBQc84PpH9R3P0oF
ur4PRJuh6Q3/GX2MF7fuNw+cweg6lNM2IlVmoH3jJo4byW+tzruv5O+/0s92CsSM
s5ywkZlgydrh1w4Irqli67y/jdDdA9zHcr+DBpVquJ1arez/ImRtKA9+FRNP4YvM
k3FOh0ECgYEA2UU+8iad7Kd7bcrhCq6AItlv51MxTp9ASDoFiCFJncTOGLdzcVNA
a+reF22XYdD32R94ldWGlIBp3MbNTyK5HYkTbwHG8414fahxg3Uy4je0NLQzHpIH
OQjaX+YFUtMDaGL7MCIDeC1FKCwfnWBRS/6xaZe3g4ne1wqZZ46DmxECgYEA5NfT
jsLSPXD5ZEz594jsOfTJ24RH4CgB69BQTd9z9AezMlTZE3fTUeXjhZRim1cs/+/4
lotnMuUEYOVRwtfJS+hqVGg1y7MFJMTo7O5RP2+SynnIrXBkZgXwKfNX0Dj4crnA
dxlUHPPFzNEZzkNMDuiwo4ERs5G+11OPD+UL6x8CgYAkaUVmQXB/44V83d4e8yWI
MZZeVwPRYEDemdKpgKKcrQm4/K19FW2baE318SjIfMO8gFiuC421P1v+YtavZ2tM
dtdp6AtWb6P8swjq9e4kGR+7IWPbwK8zMLegEKVdvv04NjZQV7LrJfMMC3D059pX
+QP0ZTec9LMCqMUSpMCLcQKBgQDGnjAnGx6AZzp9fHYECxoEX1qHpTMA8ZhhRGc+
f2/TYI9+YrgZtol57o5f1N8Utj//TxcyCoIiYTVAqCgjdUhoEque4Oe4CYOwWxtS
8LEh3sPH6pVrOz5YclT1BBi2R4wTfvb2J8yiaE3IK8A7DpvH4NvWvWJQuXGq0AI+
KG0EvwKBgB8nHRWRbNJ8admJukGb5HF2mS1tDuHi+vB1dsTydfPDyf33B1HoEG0p
mfr9uzS9ndAYCopZO33b1h65wlPP6jnIJheycn15n7HRjYezTr8cODMnJLrRotAJ
HCsYkCmGXiwJN2guZo6l/5+GqRo3SN19dZptrH/rC/wAai0+Ctqw
-----END RSA PRIVATE KEY-----`
)

func controllerNoOp(_ *Ctx) Result { return nil }

func TestAppNew(t *testing.T) {
	assert := assert.New(t)

	var route *Route
	app := New()
	assert.NotNil(app.state)
	assert.NotNil(app.Views())
	app.GET("/", func(c *Ctx) Result {
		route = c.Route()
		return c.Raw([]byte("ok!"))
	})

	assert.Nil(app.Mock().Get("/").Execute())
	assert.NotNil(route)
	assert.Equal("GET", route.Method)
	assert.Equal("/", route.Path)
	assert.NotNil(route.Handler)
}

func TestAppNewFromEnv(t *testing.T) {
	assert := assert.New(t)

	var route *Route
	app := NewFromEnv()
	assert.NotNil(app.state)
	assert.Nil(app.err)
	assert.NotNil(app.Views())
	app.GET("/", func(c *Ctx) Result {
		route = c.Route()
		return c.Raw([]byte("ok!"))
	})

	assert.Nil(app.Mock().Get("/").Execute())
	assert.NotNil(route)
	assert.Equal("GET", route.Method)
	assert.Equal("/", route.Path)
	assert.NotNil(route.Handler)
}

func TestAppNewFromConfig(t *testing.T) {
	assert := assert.New(t)

	app := NewFromConfig(&Config{
		BindAddr: ":5555",
		Port:     5000,
		HandleMethodNotAllowed: util.OptionalBool(true),
		HandleOptions:          util.OptionalBool(true),
		RecoverPanics:          util.OptionalBool(true),
		HSTS:                   util.OptionalBool(true),
		HSTSMaxAgeSeconds:      9999,
		HSTSPreload:            util.OptionalBool(false),
		HSTSIncludeSubDomains:  util.OptionalBool(false),
		MaxHeaderBytes:         128,
		ReadHeaderTimeout:      5 * time.Second,
		ReadTimeout:            6 * time.Second,
		IdleTimeout:            7 * time.Second,
		WriteTimeout:           8 * time.Second,

		CookieName: "A GOOD ONE",

		Views: ViewCacheConfig{
			Cached: util.OptionalBool(true),
		},
	})

	assert.Equal(":5555", app.BindAddr())
	assert.True(app.HandleMethodNotAllowed())
	assert.True(app.HandleOptions())
	assert.True(app.RecoverPanics())
	assert.Equal(128, app.MaxHeaderBytes())
	assert.Equal(5*time.Second, app.ReadHeaderTimeout())
	assert.Equal(6*time.Second, app.ReadTimeout())
	assert.Equal(7*time.Second, app.IdleTimeout())
	assert.Equal(8*time.Second, app.WriteTimeout())
	assert.Equal("A GOOD ONE", app.Auth().CookieName(), "we should use the auth config for the auth manager")
	assert.True(app.Views().Cached(), "we should use the view cache config for the view cache")

	assert.True(app.HSTS())
	assert.Equal(9999, app.HSTSMaxAgeSeconds())
	assert.False(app.HSTSIncludeSubdomains())
	assert.False(app.HSTSPreload())
}

func TestAppPathParams(t *testing.T) {
	assert := assert.New(t)

	var route *Route
	var params RouteParameters
	app := New()
	app.GET("/:uuid", func(c *Ctx) Result {
		route = c.Route()
		params = c.routeParameters
		return c.Raw([]byte("ok!"))
	})

	assert.Nil(app.Mock().Get("/foo").Execute())
	assert.NotNil(route)
	assert.Equal("GET", route.Method)
	assert.Equal("/:uuid", route.Path)
	assert.NotNil(route.Handler)

	assert.NotNil(params)
	assert.NotEmpty(params)
	assert.Equal("foo", params.Get("uuid"))
}

func TestAppPathParamsForked(t *testing.T) {
	assert := assert.New(t)

	var route *Route
	var params RouteParameters
	app := New()
	app.GET("/foos/bar/:uuid", func(c *Ctx) Result {
		route = c.Route()
		params = c.routeParameters
		return c.Raw([]byte("ok!"))
	})
	app.GET("/foo/:uuid", func(c *Ctx) Result { return nil })

	assert.Nil(app.Mock().Get("/foos/bar/foo").Execute())
	assert.NotNil(route)
	assert.Equal("GET", route.Method)
	assert.Equal("/foos/bar/:uuid", route.Path)
	assert.NotNil(route.Handler)

	assert.NotNil(params)
	assert.NotEmpty(params)
	assert.Equal("foo", params.Get("uuid"))
}

func TestAppSetLogger(t *testing.T) {
	assert := assert.New(t)

	log := logger.All()
	defer log.Close()
	app := New().WithLogger(log)
	assert.NotNil(app.Logger())
	assert.True(app.Logger().Flags().All())
}

func TestAppCtx(t *testing.T) {
	assert := assert.New(t)

	response := bytes.NewBuffer([]byte{})

	app := New()

	rc, err := app.Mock().WithResponseBuffer(response).Ctx(nil)
	assert.Nil(err)
	assert.NotNil(rc)
	assert.Nil(rc.log)

	result := rc.Raw([]byte("foo"))
	assert.NotNil(result)

	err = result.Render(rc)
	assert.Nil(err)
	assert.NotZero(response.Len())
	assert.True(strings.Contains(response.String(), "foo"))
}

func TestAppCreateStaticMountedRoute(t *testing.T) {
	assert := assert.New(t)
	app := New()

	assert.Equal("/testPath/*filepath", app.createStaticMountRoute("/testPath/*filepath"))
	assert.Equal("/testPath/*filepath", app.createStaticMountRoute("/testPath/"))
	assert.Equal("/testPath/*filepath", app.createStaticMountRoute("/testPath"))
}

func TestAppStaticRewrite(t *testing.T) {
	assert := assert.New(t)
	app := New()

	app.ServeStatic("/testPath", "_static")
	assert.NotEmpty(app.statics)
	assert.NotNil(app.statics["/testPath/*filepath"])
	app.WithStaticRewriteRule("/testPath", "(.*)", func(path string, pieces ...string) string {
		return path
	})
	assert.Nil(app.Err())

	assert.NotEmpty(app.statics["/testPath/*filepath"].RewriteRules())

	app.WithErr(nil)
	app.WithStaticRewriteRule("/nonsense", "(.*)", func(path string, pieces ...string) string {
		return path
	})
	assert.NotNil(app.Err())

	app.WithErr(nil)
	app.WithStaticRewriteRule("/testPath", "((((((((", func(path string, pieces ...string) string {
		return path
	})
	assert.NotNil(app.Err())
}

func TestAppStaticRewriteBadExp(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.ServeStatic("/testPath", "_static")
	assert.NotEmpty(app.statics)
	assert.NotNil(app.statics["/testPath/*filepath"])

	err := app.WithStaticRewriteRule("/testPath", "((((", func(path string, pieces ...string) string {
		return path
	})

	assert.NotNil(err)
	assert.Empty(app.statics["/testPath/*filepath"].RewriteRules())
}

func TestAppStaticHeader(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.ServeStatic("/testPath", "_static")
	assert.NotEmpty(app.statics)
	assert.NotNil(app.statics["/testPath/*filepath"])
	app.WithStaticHeader("/testPath", "cache-control", "haha what is caching.")
	assert.Nil(app.Err())
	assert.NotEmpty(app.statics["/testPath/*filepath"].Headers())

	app.WithErr(nil)
	app.WithStaticHeader("/nonsense", "cache-control", "nonsense")
	assert.NotNil(app.Err())
}

func TestAppStaticMiddleware(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.ServeStatic("/testPath", "_static")
	assert.NotEmpty(app.statics)
	assert.NotNil(app.statics["/testPath/*filepath"])
	app.WithStaticMiddleware("/testPath", func(action Action) Action {
		return func(ctx *Ctx) Result {
			return action(ctx)
		}
	})
	assert.Nil(app.Err())

	app.WithErr(nil)
	app.WithStaticMiddleware("/nonsense", func(action Action) Action {
		return func(ctx *Ctx) Result {
			return action(ctx)
		}
	})
	assert.NotNil(app.Err())
}

func TestAppMiddleWarePipeline(t *testing.T) {
	assert := assert.New(t)
	app := New()

	didRun := false
	app.GET("/",
		func(r *Ctx) Result { return r.Raw([]byte("OK!")) },
		func(action Action) Action {
			didRun = true
			return action
		},
		func(action Action) Action {
			return func(r *Ctx) Result {
				return r.Raw([]byte("foo"))
			}
		},
	)

	result, err := app.Mock().WithPathf("/").Bytes()
	assert.Nil(err)
	assert.True(didRun)
	assert.Equal("foo", string(result))
}

func TestAppStatic(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.ServeStatic("/static/*filepath", "testdata")

	index, err := app.Mock().WithPathf("/static/test_file.html").Bytes()
	assert.Nil(err)
	assert.True(strings.Contains(string(index), "Test!"), string(index))
}

func TestAppStaticSingleFile(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.GET("/", func(r *Ctx) Result {
		return r.Static("testdata/test_file.html")
	})

	index, err := app.Mock().WithPathf("/").Bytes()
	assert.Nil(err)
	assert.True(strings.Contains(string(index), "Test!"), string(index))
}

func TestAppProviderMiddleware(t *testing.T) {
	assert := assert.New(t)

	var okAction = func(r *Ctx) Result {
		assert.NotNil(r.DefaultResultProvider())
		return r.Raw([]byte("OK!"))
	}

	app := New()
	app.GET("/", okAction, JSONProviderAsDefault)

	err := app.Mock().WithPathf("/").Execute()
	assert.Nil(err)
}

func TestAppProviderMiddlewareOrder(t *testing.T) {
	assert := assert.New(t)

	app := New()

	var okAction = func(r *Ctx) Result {
		assert.NotNil(r.DefaultResultProvider())
		return r.Raw([]byte("OK!"))
	}

	var dependsOnProvider = func(action Action) Action {
		return func(r *Ctx) Result {
			assert.NotNil(r.DefaultResultProvider())
			return action(r)
		}
	}

	app.GET("/", okAction, dependsOnProvider, JSONProviderAsDefault)

	err := app.Mock().WithPathf("/").Execute()
	assert.Nil(err)
}

func TestAppDefaultResultProvider(t *testing.T) {
	assert := assert.New(t)

	app := New()
	assert.Nil(app.DefaultMiddleware())

	rc := app.createCtx(nil, nil, nil, nil, nil)
	assert.NotNil(rc.view)
	assert.NotNil(rc.json)
	assert.NotNil(rc.xml)
	assert.NotNil(rc.text)
	assert.NotNil(rc.defaultResultProvider)
}

func TestAppDefaultResultProviderWithDefault(t *testing.T) {
	assert := assert.New(t)
	app := New().WithDefaultMiddleware(ViewProviderAsDefault)
	assert.NotNil(app.DefaultMiddleware())

	rc := app.createCtx(nil, nil, nil, nil, nil)
	assert.NotNil(rc.view)
	assert.NotNil(rc.json)
	assert.NotNil(rc.xml)
	assert.NotNil(rc.text)

	// this will be set to the default initially
	assert.NotNil(rc.defaultResultProvider)

	app.GET("/", func(ctx *Ctx) Result {
		assert.NotNil(ctx.DefaultResultProvider())
		_, isTyped := ctx.DefaultResultProvider().(*ViewResultProvider)
		assert.True(isTyped)
		return nil
	})
}

func TestAppDefaultResultProviderWithDefaultFromRoute(t *testing.T) {
	assert := assert.New(t)

	app := New().WithDefaultMiddleware(JSONProviderAsDefault)
	app.Views().Templates().New(DefaultTemplateNotAuthorized).Parse("<html><body><h4>Not Authorized</h4></body></html>")
	app.GET("/", controllerNoOp, SessionRequired, ViewProviderAsDefault)

	//somehow assert that the content type is html
	meta, err := app.Mock().WithPathf("/").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(ContentTypeHTML, meta.Headers.Get(HeaderContentType))
}

func TestAppViewResult(t *testing.T) {
	assert := assert.New(t)

	app := New()
	app.Views().AddPaths("testdata/test_file.html")
	app.GET("/", func(r *Ctx) Result {
		return r.View().View("test", "foobarbaz")
	})

	res, meta, err := app.Mock().WithPathf("/").BytesWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
	assert.Equal(ContentTypeHTML, meta.Headers.Get(HeaderContentType))
	assert.Contains("foobarbaz", string(res))
}

func TestAppWritesLogs(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	agent := logger.New().WithFlags(logger.AllFlags()).WithWriter(logger.NewTextWriter(buffer))

	app := New().WithLogger(agent)
	app.GET("/", func(r *Ctx) Result {
		return r.Raw([]byte("ok!"))
	})
	err := app.Mock().Get("/").Execute()
	assert.Nil(err)
	assert.Nil(agent.Drain())

	assert.NotZero(buffer.Len())
	assert.NotEmpty(buffer.String())
}

func TestAppBindAddr(t *testing.T) {
	assert := assert.New(t)

	env.Env().Set(EnvVarBindAddr, ":9999")
	env.Env().Set(EnvVarPort, "1111")
	defer env.Restore()

	assert.Equal(":3333", New().WithBindAddr(":3333").BindAddr())
	assert.Equal(":2222", New().WithPort(2222).BindAddr())
	assert.Equal(":9999", New().WithBindAddrFromEnv().BindAddr())
	assert.Equal(":1111", New().WithPortFromEnv().BindAddr())
}

func TestAppNotFound(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	agent := logger.New().WithFlags(logger.AllFlags()).WithWriter(logger.NewTextWriter(buffer).WithShowLabel(true).WithUseColor(false).WithShowTimestamp(false))
	app := New().WithLogger(agent)
	app.GET("/", func(r *Ctx) Result {
		return r.Raw([]byte("ok!"))
	})

	wg := sync.WaitGroup{}
	wg.Add(1)

	app.WithNotFoundHandler(func(r *Ctx) Result {
		defer wg.Done()
		return r.JSON().NotFound()
	})

	agent.Listen(logger.WebRequest, "foo", logger.NewWebRequestEventListener(func(wre *logger.WebRequestEvent) {
		assert.NotNil(wre.Request())
		assert.Empty(wre.Route())
	}))

	err := app.Mock().Get("/doesntexist").Execute()
	assert.Nil(err)
	assert.Nil(agent.Drain())
	wg.Wait()
}

func TestAppDefaultHeaders(t *testing.T) {
	assert := assert.New(t)
	app := New().WithDefaultHeader("foo", "bar").WithDefaultHeader("baz", "buzz")
	app.GET("/", func(r *Ctx) Result {
		return r.Text().Result("ok")
	})

	meta, err := app.Mock().Get("/").ExecuteWithMeta()
	assert.Nil(err)
	assert.NotEmpty(meta.Headers)
	assert.Equal("bar", meta.Headers.Get("foo"))
	assert.Equal("buzz", meta.Headers.Get("baz"))
}

func TestAppIssuesHSTSHeaders(t *testing.T) {
	assert := assert.New(t)

	app := New().WithHSTS(true).WithHSTSMaxAgeSeconds(9999).WithHSTSIncludeSubdomains(true).WithHSTSPreload(true)
	app.GET("/", func(r *Ctx) Result {
		return r.Text().Result("ok")
	})
	assert.Nil(app.SetTLSCertPair([]byte(testCert), []byte(testKey)))

	meta, err := app.Mock().Get("/").ExecuteWithMeta()
	assert.Nil(err)
	assert.NotEmpty(meta.Headers)
	assert.NotEmpty(meta.Headers.Get(HeaderStrictTransportSecurity))
	assert.Equal("max-age=9999; includeSubDomains; preload", meta.Headers.Get(HeaderStrictTransportSecurity))
}

func TestAppTLSOptions(t *testing.T) {
	assert := assert.New(t)

	app := New()
	assert.NotNil(app.SetTLSCertPair([]byte{}, []byte{}))

	app = New()
	assert.Nil(app.SetTLSCertPair([]byte(testCert), []byte(testKey)))
	assert.NotNil(app.TLSConfig())
	assert.NotNil(app.TLSConfig().Certificates)

	app = New()
	assert.NotNil(app.SetTLSClientCertPool([]byte{}))
	app = New()
	assert.Nil(app.SetTLSClientCertPool([]byte(testCert)))
	assert.NotNil(app.TLSConfig())
	assert.NotNil(app.TLSConfig().ClientCAs)
	assert.NotNil(app.TLSConfig().GetConfigForClient)

	app = New()
	app.WithTLSClientCertVerification(tls.RequireAndVerifyClientCert)
	assert.NotNil(app.TLSConfig())
	assert.Equal(tls.RequireAndVerifyClientCert, app.TLSConfig().ClientAuth)

	app = New()
	app.WithTLSClientCertVerification(tls.RequireAndVerifyClientCert)
	assert.Nil(app.SetTLSCertPair([]byte(testCert), []byte(testKey)))
	assert.NotNil(app.TLSConfig())
	assert.NotNil(app.TLSConfig().Certificates)
	assert.Equal(tls.RequireAndVerifyClientCert, app.TLSConfig().ClientAuth)
}

func TestAppTLSCertPair(t *testing.T) {
	assert := assert.New(t)
	app := New().WithTLSCertPair([]byte(testCert), []byte(testKey))
	assert.Nil(app.Err())
	assert.NotNil(app.TLSConfig())
	assert.NotEmpty(app.TLSConfig().Certificates)
}

func TestAppTLSCertsFromFiles(t *testing.T) {
	assert := assert.New(t)
	app := New().WithTLSCertPairFromFiles("testdata/testcert.pem", "testdata/testkey.pem")
	assert.Nil(app.Err())
	assert.NotNil(app.TLSConfig())
	assert.NotEmpty(app.TLSConfig().Certificates)
}

func TestAppTLSCertsFromFilesInvalid(t *testing.T) {
	assert := assert.New(t)
	app := New().WithTLSCertPairFromFiles(util.String.Random(32), util.String.Random(32))
	assert.NotNil(app.Err())
}

func TestAppTLSCertsFromEnv(t *testing.T) {
	assert := assert.New(t)
	defer env.Restore()
	env.Env().Delete(EnvVarTLSCert)
	env.Env().Delete(EnvVarTLSKey)
	env.Env().Set(EnvVarTLSCertPath, "testdata/testcert.pem")
	env.Env().Set(EnvVarTLSKeyPath, "testdata/testkey.pem")
	app := New().WithTLSFromEnv()
	assert.Nil(app.Err())
	assert.NotNil(app.TLSConfig())
	assert.NotEmpty(app.TLSConfig().Certificates)
}

func TestAppProperties(t *testing.T) {
	assert := assert.New(t)

	app := New()

	assert.Nil(app.Err())
	app.WithErr(Error("this is only a test"))
	assert.Equal("this is only a test", app.Err())

	assert.Empty(app.State())
	app.WithState("foo", "bar-state")
	assert.Equal("bar-state", app.State()["foo"])

	assert.NotNil(app.DefaultHeaders())
	app.WithDefaultHeader("foo", "bar-header")
	assert.Equal("bar-header", app.DefaultHeaders()["foo"])

	assert.Equal(DefaultRedirectTrailingSlash, app.RedirectTrailingSlash())
	app.WithRedirectTrailingSlash(!DefaultRedirectTrailingSlash)
	assert.Equal(!DefaultRedirectTrailingSlash, app.RedirectTrailingSlash())

	assert.Nil(app.BaseURL())
	app.WithBaseURL(&url.URL{Host: "localhost"})
	assert.Equal("localhost", app.BaseURL().Host)

	app.WithErr(nil)
	app.WithParsedBaseURL(":foo-parsed")
	assert.NotNil(app.Err())
	app.WithErr(nil)
	app.WithParsedBaseURL("https://localhost-parsed")
	assert.Equal("https", app.BaseURL().Scheme)
	assert.Equal("localhost-parsed", app.BaseURL().Host)

	assert.NotNil(app.Server())

	assert.NotNil(app.ViewResultProvider())
	app.WithViewResultProvider(nil)
	assert.Nil(app.ViewResultProvider())

	assert.NotNil(app.JSONResultProvider())
	app.WithJSONResultProvider(nil)
	assert.Nil(app.JSONResultProvider())

	assert.NotNil(app.XMLResultProvider())
	app.WithXMLResultProvider(nil)
	assert.Nil(app.XMLResultProvider())

	assert.NotNil(app.TextResultProvider())
	app.WithTextResultProvider(nil)
	assert.Nil(app.TextResultProvider())

	assert.Nil(app.DefaultResultProvider())
	app.WithDefaultResultProvider(&TextResultProvider{})
	assert.NotNil(app.DefaultResultProvider())
}

func TestAppStartWithError(t *testing.T) {
	assert := assert.New(t)

	app := New().WithErr(Error("this is only a test"))
	assert.Equal("this is only a test", app.Start(), "we should exit early if there is a prestart error")
}

func TestAppStartDelegateWithError(t *testing.T) {
	assert := assert.New(t)

	app := New().WithOnStart(func(_ *App) error {
		return Error("this is only a test")
	})
	assert.Equal("this is only a test", app.Start(), "we should exit early if there is a prestart error")
}
