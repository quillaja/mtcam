// package main starts a server to respond to api requests for scrape info.
// It also serves the client application and (if configured) will redirect
// HTTP requests to HTTPS.
package main

import (
	"context"
	"flag"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quillaja/mtcam/config"
	"github.com/quillaja/mtcam/db"
	"github.com/quillaja/mtcam/log"
	"github.com/quillaja/mtcam/version"
)

//go:generate go run generate_client.go "../../client"

func main() {

	// process command line flags
	configPath := flag.String("cfg", "", "path to served config (required)\n'default' to produce a default config file")
	flag.Usage = func() {
		fmt.Print("served is the web front-end server for mountain cameras.\n\n")
		fmt.Printf("Version:  %s\nBuilt on: %s\n\nOptions:\n", version.Version, version.BuildTime)
		flag.PrintDefaults()
	}
	flag.Parse()

	switch *configPath {
	case "":
		flag.Usage()
		return
	case "default":
		config.Write("served_config_default.json", ServerdConfig{})
		config.Write("suite_config_default.json", config.SuiteConfig{})
		return
	}

	// read and create config
	var cfg ServerdConfig
	err := config.Read(*configPath, &cfg)
	if err != nil {
		log.Printf(log.Error, "could read config %s: %s", *configPath, err)
		return
	}
	err = config.Read(cfg.SuiteConfigPath, &cfg.SuiteConfig)
	if err != nil {
		log.Printf(log.Error, "could read suit config %s: %s", cfg.SuiteConfigPath, err)
		return
	}
	// TODO: watch config file(s) and update on the fly.

	// 'connect' to database
	err = db.Connect(cfg.DatabaseConnection)
	defer db.Close()
	if err != nil {
		log.Printf(log.Error, "error connecting to db: %s", err)
		return
	}

	app := NewApplication(&cfg)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Kill, os.Interrupt, syscall.SIGTERM)

	log.Printf(log.Info, "starting server daemon %s", version.Version)
	app.run()

	<-sig // wait for SIGKILL, SIGINT, or SIGTERM

	log.Printf(log.Info, "shutting down server daemon %s", version.Version)
	app.shutdown()
	log.Print(log.Info, "server daemon finished")
}

// Application is the served application logic.
type Application struct {
	Config      *ServerdConfig
	HttpsServer *http.Server
	HttpServer  *http.Server
}

// NewApplication configures and returns an instance of Application.
func NewApplication(cfg *ServerdConfig) *Application {
	app := Application{
		Config: cfg}

	// set default addresses
	if cfg.HttpsAddress == "" {
		cfg.HttpsAddress = ":https"
	}
	if cfg.HttpAddress == "" {
		cfg.HttpAddress = ":http"
	}

	dohttps := cfg.TLSCertificateFile != "" && cfg.TLSKeyFile != ""

	// configure HTTPS server as main and HTTP as redirect
	// or HTTP as main, depending on if TLS files are provided
	if dohttps {
		app.HttpsServer = &http.Server{
			Addr:         cfg.HttpsAddress,
			IdleTimeout:  time.Duration(cfg.Timeout.Idle) * time.Second,
			ReadTimeout:  time.Duration(cfg.Timeout.Read) * time.Second,
			WriteTimeout: time.Duration(cfg.Timeout.Write) * time.Second,
			Handler:      CreateHandler(cfg),
			ErrorLog:     stdlog.New(serverlogwriter{}, "HTTPS ", stdlog.Lshortfile),
		}

		app.HttpServer = redirectHTTPS(cfg)

		log.Printf(log.Info, "listening on %s, redirecting from %s",
			cfg.HttpsAddress, cfg.HttpAddress)
		if len(cfg.RedirectedHosts) > 0 {
			log.Printf(log.Info, "redirecting hosts %v", cfg.RedirectedHosts)
		}

	} else {
		app.HttpServer = &http.Server{
			Addr:         cfg.HttpAddress,
			IdleTimeout:  time.Duration(cfg.Timeout.Idle) * time.Second,
			ReadTimeout:  time.Duration(cfg.Timeout.Read) * time.Second,
			WriteTimeout: time.Duration(cfg.Timeout.Write) * time.Second,
			Handler:      CreateHandler(cfg),
			ErrorLog:     stdlog.New(serverlogwriter{}, "HTTP ", stdlog.Lshortfile),
		}

		log.Printf(log.Info, "listening on %s", cfg.HttpAddress)
	}

	return &app
}

// run starts each server
func (app *Application) run() {
	if app.HttpsServer != nil {
		go func() {
			err := app.HttpsServer.ListenAndServeTLS(
				app.Config.TLSCertificateFile,
				app.Config.TLSKeyFile)
			if err != nil && err != http.ErrServerClosed {
				// unexpected error
				log.Printf(log.Critical, "error with https server: %s", err)
			}
		}()
	}

	go func() {
		err := app.HttpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			// unexpected error
			log.Printf(log.Critical, "error with http server: %s", err)
		}
	}()
}

// shutdown attempts to shutdown each server, waiting up to 30 sec.
func (app *Application) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := app.HttpServer.Shutdown(ctx)
	if err != nil {
		log.Printf(log.Error, "error shutting down http server: %s", err)
	}

	if app.HttpsServer != nil {
		err = app.HttpsServer.Shutdown(ctx)
		if err != nil {
			log.Printf(log.Error, "error shutting down https server: %s", err)
		}
	}
}

// listens on :http and redirects to the same address, just with https.
func redirectHTTPS(cfg *ServerdConfig) *http.Server {

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// deny http request if they're asking for anything but / on one of
		// the specified hosts.
		if r.RequestURI != "/" ||
			(len(cfg.RedirectedHosts) > 0 && !hostInList(r.Host, cfg.RedirectedHosts)) {
			//log.Printf(log.Debug, "%s denied redirect host: %s uri: %s", r.RemoteAddr, r.Host, r.RequestURI)
			http.NotFound(w, r)
			return
		}
		nurl := *r.URL
		nurl.Scheme = "https"
		nurl.Host = r.Host
		// put a new port on host if the https server is on a non-standard port
		if !(cfg.HttpsAddress == ":https" || cfg.HttpsAddress == ":443") {
			nurl.Host = net.JoinHostPort(nurl.Hostname(), cfg.HttpsAddress[1:]) // a little hackish. assumes colon.
		}
		log.Printf(log.Debug, "%s redirecting %s to %s", r.RemoteAddr, r.RequestURI, nurl.String())
		http.Redirect(w, r, nurl.String(), http.StatusPermanentRedirect)
	})

	srv := http.Server{
		Addr:         cfg.HttpAddress,
		IdleTimeout:  time.Duration(cfg.Timeout.Idle) * time.Second,
		ReadTimeout:  time.Duration(cfg.Timeout.Read) * time.Second,
		WriteTimeout: time.Duration(cfg.Timeout.Write) * time.Second,
		Handler:      handler,
		ErrorLog:     stdlog.New(serverlogwriter{}, "HTTP-REDIRECT ", stdlog.Lshortfile),
	}

	return &srv
}

func hostInList(host string, list []string) bool {
	h, _, err := net.SplitHostPort(host)
	if err == nil {
		host = h
	}
	for _, h := range list {
		if host == h {
			return true
		}
	}
	return false
}

// used to "forward" the server's internal logging to the application's
// systemd log system
type serverlogwriter struct{}

func (w serverlogwriter) Write(p []byte) (n int, err error) {
	log.Printf(log.Debug, "  %s", string(p))
	return len(p), nil
}
