package main

import "github.com/quillaja/mtcam/config"

type ServerdConfig struct {
	config.SuiteConfig `json:"-"`

	SuiteConfigPath string // path to suite config file

	HttpsAddress string // eg 123.1.1.123:8080, :8080, :http etc
	HttpAddress  string

	// TLS (https) certificate stuff
	TLSCertificateFile string
	TLSKeyFile         string

	// various timeouts for server, in seconds
	WriteTimeout int
	ReadTimeout  int
	IdleTimeout  int

	// static root directory
	StaticRoot string
}
