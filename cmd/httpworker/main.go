package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/jlym/webservice-benchmarks/sqlite"
	"github.com/jlym/webservice-benchmarks/util"
	"github.com/pkg/errors"
)

func main() {
	runTest(&testConfig{
		numWorkers:     2,
		dbFilePath:     "data3.sqlite3",
		rampUpDuration: time.Second * 10,
		testDuration:   time.Second * 20,
		port:           8080,
		runID:          util.NewID(),
	})
}

type testConfig struct {
	dbFilePath     string
	numWorkers     int
	rampUpDuration time.Duration
	testDuration   time.Duration
	port           int
	runID          string
}

func runTest(config *testConfig) {
	ctx := context.Background()
	c := newClient(fmt.Sprintf("localhost:%d", config.port))

	data, err := sqlite.NewDataStore(config.dbFilePath)
	if err != nil {
		log.Fatal(err)
	}
	err = data.CreateTables(ctx)
	if err != nil {
		log.Fatal(err)
	}
	data.Start()
	defer func() {
		data.Stop()
		err := data.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	run := &sqlite.Run{
		ID:        config.runID,
		StartTime: time.Now().UTC(),
	}
	err = data.WriteRunStart(ctx, &sqlite.AddRunParams{
		ID:         run.ID,
		StartTime:  run.StartTime,
		NumWorkers: config.numWorkers,
	})
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	stopper := newStopper()

	for workerID := 0; workerID < config.numWorkers; workerID++ {
		log.Println("start: ", workerID)
		w := newWorker(c, workerID, data, stopper, run, wg)
		w.Start(ctx)
		wg.Add(1)

		if workerID < config.numWorkers-1 {
			ch := time.After(config.rampUpDuration)
			<-ch
		}
	}

	ch := time.After(config.testDuration)
	<-ch

	stopper.stop()
	wg.Wait()

	log.Println("Done")
}

type worker struct {
	id int
	c  *client

	data *sqlite.DataStore
	run  *sqlite.Run

	stopper *stopper
	wg      *sync.WaitGroup
}

func newWorker(c *client, id int, data *sqlite.DataStore, stopper *stopper, run *sqlite.Run, wg *sync.WaitGroup) *worker {
	return &worker{
		id: id,
		c:  c,

		data:    data,
		stopper: stopper,
		run:     run,
		wg:      wg,
	}
}

func (w *worker) Start(ctx context.Context) {
	go func() {
		w.doRun(ctx)
	}()
}

func (w *worker) doRun(ctx context.Context) {
	defer w.wg.Done()

	for w.stopper.shouldContinue() {
		start := time.Now()
		_, err := w.c.calcNthPrime(1000)
		end := time.Now()

		errorMessage := ""
		if err != nil {
			errorMessage = err.Error()
		}

		w.data.QueueClientRequest(w.run, &sqlite.AddRequestParams{
			WorkerID:  w.id,
			StartTime: start,
			EndTime:   end,
			Success:   err == nil,
			Error:     errorMessage,
		})
	}
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

type stopper struct {
	done chan struct{}
}

func newStopper() *stopper {
	return &stopper{
		done: make(chan struct{}),
	}
}

func (r *stopper) stop() {
	close(r.done)
}

func (r *stopper) shouldStop() bool {
	select {
	case <-r.done:
		return true
	default:
		return false
	}
}

func (r *stopper) shouldContinue() bool {
	return !r.shouldStop()
}
