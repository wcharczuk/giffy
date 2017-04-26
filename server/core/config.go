package core

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/env"
	"github.com/blendlabs/spiffy"
)

// DBInit reads the config from the environment and sets up spiffy.
func DBInit() error {
	err := spiffy.OpenDefault(spiffy.NewConnectionFromEnvironment())
	if err != nil {
		return err
	}
	spiffy.Default().Connection.SetMaxIdleConns(50)
	return nil
}

// ConfigPort is the port the server should listen on.
func ConfigPort() string {
	envPort := os.Getenv("PORT")
	if !util.String.IsEmpty(envPort) {
		return envPort
	}
	return "8080"
}

// ConfigLocalIP is the server local IP.
func ConfigLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// ConfigKey is the app secret we use to encrypt things.
func ConfigKey() []byte {
	return env.Env().Base64("ENCRYPTION_KEY")
}

// ConfigEnvironment returns the current environment.
func ConfigEnvironment() string {
	env := os.Getenv("GIFFY_ENV")
	if len(env) != 0 {
		return env
	}
	return "dev"
}

// ConfigIsProduction returns if the app is running in production mode.
func ConfigIsProduction() bool {
	return ConfigEnvironment() == "prod"
}

// ConfigHostname returns the hostname for the server.
func ConfigHostname() string {
	envHost := os.Getenv("HOSTNAME")
	if len(envHost) != 0 {
		return envHost
	}

	if ConfigIsProduction() {
		return "giffy.charczuk.com"
	}

	return "dev.giffy.charczuk.com"
}

// ConfigHTTPProto is the proto for the webserver.
func ConfigHTTPProto() string {
	envProto := os.Getenv("PROTO")
	if len(envProto) != 0 {
		return envProto
	}

	if ConfigIsProduction() {
		return "https"
	}
	return "http"
}

// ConfigURL is the url root for the server.
func ConfigURL() string {
	return fmt.Sprintf("%s://%s", ConfigHTTPProto(), ConfigHostname())
}

// ConfigGoogleClientID returns the google client id.
func ConfigGoogleClientID() string {
	return os.Getenv("GOOGLE_CLIENT_ID")
}

// ConfigGoogleSecret returns the google secret.
func ConfigGoogleSecret() string {
	return os.Getenv("GOOGLE_CLIENT_SECRET")
}

// ConfigSlackClientID is the verification token we use for slack requests.
func ConfigSlackClientID() string {
	return os.Getenv("SLACK_CLIENT_ID")
}

// ConfigSlackClientSecret is the verification token we use for slack requests.
func ConfigSlackClientSecret() string {
	return os.Getenv("SLACK_CLIENT_SECRET")
}

// ConfigStathatToken returns the stathat token.
func ConfigStathatToken() string {
	return os.Getenv("STATHAT_TOKEN")
}

// ConfigFacebookClientID returns the facebook client id.
func ConfigFacebookClientID() string {
	return os.Getenv("FACEBOOK_CLIENT_ID")
}

// ConfigFacebookClientSecret returns the bacebook client secret.
func ConfigFacebookClientSecret() string {
	return os.Getenv("FACEBOOK_CLIENT_SECRET")
}

// Setwd sets the working directory to the relative path.
func Setwd(relativePath string) {
	fullPath, err := filepath.Abs(relativePath)
	if err != nil {
		return
	}
	os.Chdir(fullPath)
}
