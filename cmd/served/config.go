package main

import "github.com/quillaja/mtcam/config"

// ServedConfig holds configuration settings for the served program.
type ServerdConfig struct {
	config.SuiteConfig `json:"-"`

	SuiteConfigPath string // path to suite config file

	HttpsAddress string // eg 123.1.1.123:8080, :8080, :http etc
	HttpAddress  string

	// TLS (https) certificate stuff
	TLSCertificateFile string
	TLSKeyFile         string

	Timeout TimeoutConfig

	// static root directory
	StaticRoot string

	// api and image routes
	Routes RoutesConfig

	RedirectedHosts []string

	// file path for json log of scrape requests
	RequestLog string
}

// TimeoutConfig holds timeouts for server, in seconds.
type TimeoutConfig struct {
	Write int
	Read  int
	Idle  int
}

// RoutesConfig holds url routes.
type RoutesConfig struct {
	Api   string
	Image string
}
