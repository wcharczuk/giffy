package core

import (
	"net"
	"os"
	"path/filepath"

	exception "github.com/blend/go-sdk/ex"
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
