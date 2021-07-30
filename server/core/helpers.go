package core

import (
	"net"
	"os"
	"path/filepath"

	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/exception"
	"github.com/wcharczuk/giffy/server/config"
)

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

// Setwd sets the working directory to the relative path.
func Setwd(relativePath string) error {
	fullPath, err := filepath.Abs(relativePath)
	if err != nil {
		return exception.New(err)
	}
	return exception.New(os.Chdir(fullPath))
}

// InitTest initializes the test prereqs.
func InitTest() error {
	var cfg config.Giffy
	if err := configutil.Read(&cfg); !configutil.IsIgnored(err) {
		return err
	}

	return db.OpenDefault(db.MustNewFromConfig(&cfg.DB))
}
