// Package main implements a scraping daemon.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
	taskwait := 30 * time.Second
	app := &Application{
		Config: &cfg,
		Scheduler: scheduler.NewScheduler(
			scheduler.WaitForUnfinishedTasks(taskwait)),
	}

	log.Printf(log.Info, "starting scrape daemon %s", version.Version)
	err = app.run()
	if err != nil {
		log.Printf(log.Critical, "error running application: %s", err)
		return
	}

	// wait for os signals to end app
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Kill, os.Interrupt, syscall.SIGTERM)

	<-sig

	log.Printf(log.Info, "waiting %s for %d tasks to complete...", taskwait, app.Scheduler.Running())
	app.shutdown() // will wait 30 sec for unfinished tasks to complete

	log.Printf(log.Info, "shutting down scrape daemon %s", version.Version)
}

// Application is the scraped app.
type Application struct {
	Config    *ScrapedConfig
	Scheduler *scheduler.Scheduler

	cancel context.CancelFunc
}

// run starts the scheduler, adds tasks to schedule scrapes, and blocks.
func (app *Application) run() error {
	// start scheduler
	var ctx context.Context
	ctx, app.cancel = context.WithCancel(context.Background())
	app.Scheduler.Start(ctx)

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

	return nil
}

func (app *Application) shutdown() {
	app.cancel()

	// block on scheduler
	app.Scheduler.Wait()
}
