package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
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
			Name:     "target-process-names",
			Usage:    "The names of the processes that the test involves.",
			Required: true,
		},
		cli.BoolFlag{
			Name:  "include-all-processes",
			Usage: "Record stats on all processes.",
		},
		cli.DurationFlag{
			Name:  "startup-wait",
			Usage: "The maximum amount of time monitor will wait for the target processes to start.",
			Value: time.Minute * 2,
		},
		cli.DurationFlag{
			Name:  "shutdown-wait",
			Usage: "The amount of time the monitor waits after the target processes exited.",
			Value: time.Minute * 2,
		},
		cli.DurationFlag{
			Name:  "polling-interval",
			Usage: "The amount of time monitor waits in between getting information on the processes.",
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

	log.Println(fmt.Sprintf("runID: %v", conf.runID))
	log.Println(fmt.Sprintf("dbFilePath: %v", conf.dbFilePath))
	log.Println(fmt.Sprintf("targetProcessNames: %v", strings.Join(conf.targetProcessNames, ", ")))
	log.Println(fmt.Sprintf("includeAllProcesses: %v", conf.includeAllProcesses))
	log.Println(fmt.Sprintf("startupWait: %v", conf.startupWait))
	log.Println(fmt.Sprintf("shutdownWait: %v", conf.shutdownWait))
	log.Println(fmt.Sprintf("pollingInterval: %v", conf.pollingInterval))

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
