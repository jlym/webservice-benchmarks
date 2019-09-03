package main

import (
	"fmt"
	"log"
	"time"

	webservice_benchmarks "github.com/jlym/webservice-benchmarks"
	"github.com/jlym/webservice-benchmarks/util"

	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

func main() {
	runTest(&webservice_benchmarks.TestConfig{
		NumWorkers:     2,
		DBFilePath:     "data4.sqlite3",
		RampUpDuration: time.Second * 10,
		TestDuration:   time.Second * 20,
		Port:           8080,
		RunID:          util.NewID(),
	})
}

func runTest(config *webservice_benchmarks.TestConfig) {
	client := newClient(fmt.Sprintf("localhost:%d", config.Port))

	err := webservice_benchmarks.GenerateLoad(config, func(workerID int) error {
		_, err := client.calcNthPrime(1000)
		return err
	})
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
