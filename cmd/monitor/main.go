package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jlym/webservice-benchmarks/sqlite"
	"github.com/jlym/webservice-benchmarks/util"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:     "db",
			Usage:    "Path to sqlite database that results should be written to.",
			Required: true,
		},
		cli.StringFlag{
			Name:  "run-id",
			Usage: "Path to sqlite database that results should be written to.",
		},
		cli.StringSliceFlag{
			Name:  "target-process-names",
			Usage: "Path to sqlite database that results should be written to.",
		},
		cli.BoolFlag{
			Name:  "include-all-processes",
			Usage: "Path to sqlite database that results should be written to.",
		},
		cli.DurationFlag{
			Name:  "startup-wait",
			Usage: "The amount of time the test should run (after ramp up).",
			Value: time.Minute * 2,
		},
		cli.DurationFlag{
			Name:  "shutdown-wait",
			Usage: "The amount of time the test waits before creating a new goroutine.",
			Value: time.Minute * 2,
		},
		cli.DurationFlag{
			Name:  "polling-interval",
			Usage: "The amount of time the test waits before creating a new goroutine.",
			Value: time.Second,
		},
	}
	app.Action = func(c *cli.Context) error {
		conf := getConfig(c)
		return run(conf)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Done")
}

type config struct {
	dbFilePath          string
	runID               string
	targetProcessNames  []string
	includeAllProcesses bool
	startupWait         time.Duration
	shutdownWait        time.Duration
	pollingInterval     time.Duration
}

func getConfig(c *cli.Context) *config {
	runID := c.String("run-id")
	if runID == "" {
		runID = util.NewID()
	}

	return &config{
		runID:               runID,
		dbFilePath:          c.String("db"),
		targetProcessNames:  c.StringSlice("target-process-names"),
		includeAllProcesses: c.Bool("include-all-processes"),
		startupWait:         c.Duration("startup-wait"),
		shutdownWait:        c.Duration("shutdown-wait"),
		pollingInterval:     c.Duration("polling-interval"),
	}
}

func run(conf *config) error {
	ctx := context.Background()

	db, err := sqlite.NewDataStore(conf.dbFilePath)
	if err != nil {
		return err
	}

	err = db.CreateTables(ctx)
	if err != nil {
		return err
	}

	db.Start()
	defer func() {
		db.Stop()
		_ = db.Close()
	}()

	return monitor(conf, db)
}
