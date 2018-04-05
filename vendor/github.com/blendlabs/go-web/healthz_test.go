package web

import (
	"net/http"
	"testing"

	assert "github.com/blendlabs/go-assert"
	logger "github.com/blendlabs/go-logger"
)

func TestHealthz(t *testing.T) {
	assert := assert.New(t)

	appLog := logger.New().WithFlags(logger.AllFlags())
	defer appLog.Close()

	app := New().WithBindAddr("127.0.0.1:0").WithLogger(appLog)
	defer app.Shutdown()

	appStarted := make(chan struct{})
	appLog.Listen(AppStartComplete, "default", NewAppEventListener(func(aes *AppEvent) {
		close(appStarted)
	}))

	hzLog := logger.New().WithFlags(logger.AllFlags())
	defer hzLog.Close()
	hz := NewHealthz(app).WithBindAddr("127.0.0.1:0").WithLogger(hzLog)
	defer hz.Shutdown()

	hzStarted := make(chan struct{})
	hzLog.Listen(HealthzStartComplete, "default", NewAppEventListener(func(aes *AppEvent) {
		close(hzStarted)
	}))

	assert.NotNil(hz.App())
	assert.False(app.Running())

	go app.Start()
	go hz.Start()

	<-appStarted
	<-hzStarted

	assert.True(app.Running())
	assert.True(hz.App().Running())

	assert.NotNil(hz.Listener())

	healthzRes, err := http.Get("http://" + hz.Listener().Addr().String() + "/healthz")
	assert.Nil(err)
	assert.Equal(http.StatusOK, healthzRes.StatusCode)
}
