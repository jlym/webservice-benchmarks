package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	webservice_benchmarks "github.com/jlym/webservice-benchmarks"
	"github.com/jlym/webservice-benchmarks/util"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	config := &webservice_benchmarks.TestConfig{
		RunID: util.NewID(),
	}
	var serverBaseEndpoint string

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "db",
			Usage:       "Path to sqlite database that results should be written to.",
			Required:    true,
			Destination: &config.DBFilePath,
		},
		cli.IntFlag{
			Name:        "parallel",
			Usage:       "Number of goroutines that will be sending requests.",
			Value:       10,
			Destination: &config.NumWorkers,
		},
		cli.DurationFlag{
			Name:        "test-duration",
			Usage:       "The amount of time the test should run (after ramp up).",
			Value:       time.Minute * 2,
			Destination: &config.TestDuration,
		},
		cli.DurationFlag{
			Name:        "ramp-up-duration",
			Usage:       "The amount of time the test waits before creating a new goroutine.",
			Value:       0,
			Destination: &config.RampUpDuration,
		},
		cli.StringFlag{
			Name:        "endpoint",
			Usage:       "The base endpoint of the server. (format: [hostname]:[port])",
			Value:       "localhost:8080",
			Destination: &serverBaseEndpoint,
		},
	}
	app.Action = func(_ *cli.Context) error {
		log.Println(fmt.Sprintf("DBFilePath: %v", config.DBFilePath))
		log.Println(fmt.Sprintf("NumWorkers: %v", config.NumWorkers))
		log.Println(fmt.Sprintf("RampUpDuration: %v", config.RampUpDuration))
		log.Println(fmt.Sprintf("RunID: %v", config.RunID))
		log.Println(fmt.Sprintf("TestDuration: %v", config.TestDuration))
		log.Println(fmt.Sprintf("ServiceBaseEndpoint: %v", serverBaseEndpoint))

		client := newClient(serverBaseEndpoint)

		return webservice_benchmarks.GenerateLoad(config, func(workerID int) error {
			_, err := client.calcNthPrime(1000)
			return err
		})
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Done")
}

type client struct {
	client       *http.Client
	baseEndpoint string
}

func newClient(baseEndpoint string) *client {
	return &client{
		client:       &http.Client{},
		baseEndpoint: baseEndpoint,
	}
}

func (c *client) calcNthPrime(n int) (int, error) {
	query := url.Values{}
	query.Set("n", strconv.Itoa(n))

	u := url.URL{
		Scheme:   "http",
		Host:     c.baseEndpoint,
		Path:     "prime",
		RawQuery: query.Encode(),
	}
	resp, err := c.client.Get(u.String())
	if err != nil {
		return 0, errors.Wrap(err, "sending request failed")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.Wrap(err, "reading response body failed")
	}

	if resp.StatusCode != http.StatusOK {
		return 0, errors.Errorf("unexpected stauts - %d %s; body - %s", resp.StatusCode, resp.Status, string(body))
	}

	result, err := strconv.Atoi(string(body))
	if err != nil {
		return 0, errors.Wrap(err, "response body was not an int")
	}

	return result, nil
}
