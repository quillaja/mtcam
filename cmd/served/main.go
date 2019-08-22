package main

import (
	"context"
	"flag"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/quillaja/mtcam/config"
	"github.com/quillaja/mtcam/db"
	"github.com/quillaja/mtcam/log"
)

//go:generate go run generate_client.go "../../client"

func main() {

	// process command line flags
	configPath := flag.String("cfg", "", "path to served config (required)")
	flag.Parse()

	if *configPath == "" {
		flag.Usage()
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
	signal.Notify(sig, os.Kill, os.Interrupt)

	log.Print(log.Info, "starting server daemon")
	app.run()

	<-sig // wait for SIGKILL or SIGINT

	log.Print(log.Info, "shutting down server daemon")
	app.shutdown()
	log.Print(log.Info, "server daemon finished")
}

type Application struct {
	Config      *ServerdConfig
	HttpsServer *http.Server
	HttpServer  *http.Server
}

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
			IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
			Handler:      CreateHandler(cfg),
			ErrorLog:     stdlog.New(serverlogwriter{}, "HTTPS ", stdlog.Lshortfile),
		}

		app.HttpServer = redirectHTTPS(cfg)

		log.Printf(log.Info, "listening on %s, redirecting from %s",
			cfg.HttpsAddress, cfg.HttpAddress)

	} else {
		app.HttpServer = &http.Server{
			Addr:         cfg.HttpAddress,
			IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
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
	mux := new(http.ServeMux)

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nurl := *r.URL
		nurl.Scheme = "https"
		nurl.Host = r.Host
		log.Printf(log.Info, "redirecting http to %s", nurl.String())
		http.Redirect(w, r, nurl.String(), http.StatusPermanentRedirect)
	}))

	srv := http.Server{
		Addr:         cfg.HttpAddress,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		Handler:      mux,
		ErrorLog:     stdlog.New(serverlogwriter{}, "HTTP-REDIRECT ", stdlog.Lshortfile),
	}

	return &srv
}

// used to "forward" the server's internal logging to the application's
// systemd log system
type serverlogwriter struct{}

func (w serverlogwriter) Write(p []byte) (n int, err error) {
	log.Print(log.Error, string(p))
	return len(p), nil
}
