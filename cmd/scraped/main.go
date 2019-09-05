// Package main implements a scraping daemon.
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/quillaja/mtcam/config"
	"github.com/quillaja/mtcam/db"
	"github.com/quillaja/mtcam/log"
	"github.com/quillaja/mtcam/scheduler"
	"github.com/quillaja/mtcam/version"
)

func main() {

	// process command line flags
	configPath := flag.String("cfg", "", "path to scraped config (required)\n'default' to produce a default config file")
	flag.Usage = func() {
		fmt.Print("scraped is the web scraping daemon for mountain cameras.\n\n")
		fmt.Printf("Version:  %s\nBuilt on: %s\n\nOptions:\n", version.Version, version.BuildTime)
		flag.PrintDefaults()
	}
	flag.Parse()

	switch *configPath {
	case "":
		flag.Usage()
		return
	case "default":
		config.Write("scraped_config_default.json", ScrapedConfig{})
		config.Write("suite_config_default.json", config.SuiteConfig{})
		return
	}

	// read and create config
	var cfg ScrapedConfig
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
	// TODO: watch config file(s) and update config on the fly.
	// What about options that can't be (easily) changed once the program is
	// running, such as DB connection, and config watch interval?

	// 'connect' to database
	err = db.Connect(cfg.DatabaseConnection)
	defer db.Close()
	if err != nil {
		log.Printf(log.Error, "error connecting to db: %s", err)
		return
	}

	// initialize application and run
	app := &Application{
		Config:    &cfg,
		Scheduler: scheduler.NewScheduler()}

	log.Printf(log.Info, "starting scrape daemon %s", version.Version)
	err = app.run()
	if err != nil {
		log.Printf(log.Critical, "error running application: %s", err)
		return
	}
}

// Application is the scraped app.
type Application struct {
	Config    *ScrapedConfig
	Scheduler *scheduler.Scheduler
}

// run starts the scheduler, adds tasks to schedule scrapes, and blocks.
func (app *Application) run() error {
	// start scheduler
	app.Scheduler.Start(context.Background())

	// load scheduler with some tasks
	mts, err := db.Mountains()
	if err != nil {
		return errors.Wrap(err, "reading db in app.run()")
	}
	for id := range mts {
		app.Scheduler.Add(scheduler.NewTask(
			time.Now(),
			ScheduleScrapes(id, 0, app)))
	}

	// block on scheduler
	app.Scheduler.Wait()
	return nil
}
