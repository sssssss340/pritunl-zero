package constants

import (
	"path/filepath"
	"time"
)

const (
	Version         = "1.0.619.29"
	DatabaseVersion = 1
	ConfPath        = "/etc/pritunl-zero.json"
	LogPath         = "/var/log/pritunl-zero.log"
	TempPath        = "/tmp/pritunl-zero"
	StaticCache     = true
	RetryDelay      = 3 * time.Second
)

var (
	Production = true
	Interrupt  = false
	StaticRoot = []string{
		"www/dist",
		"/usr/share/pritunl-zero/www",
	}
	StaticTestingRoot = []string{
		"www",
		"/usr/share/pritunl-zero/www",
	}
	CertPath = filepath.Join(TempPath, "server.crt")
	KeyPath  = filepath.Join(TempPath, "server.key")
)
