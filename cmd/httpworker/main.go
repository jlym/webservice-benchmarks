package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jlym/webservice-benchmarks/sqlite"
	"github.com/pkg/errors"
)

func main() {
	c := newClient("localhost:8080")

	data, err := sqlite.NewDataStore("./data.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	err = data.CreateTables(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	data.Start()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	for i := 0; i < 10; i++ {
		log.Println("start: ", i)
		w := newWorker(c, i, data)
		w.Start(ctx)
	}

	<-ctx.Done()
	log.Println("Done")

}

type worker struct {
	id   int
	c    *client
	done chan struct{}
	data *sqlite.DataStore
	run  *sqlite.Run
}

func newWorker(c *client, id int, data *sqlite.DataStore) *worker {
	return &worker{
		id:   id,
		c:    c,
		done: make(chan struct{}),
		data: data,
	}
}

func (w *worker) Start(ctx context.Context) {
	go func() {
		w.doRun(ctx)
	}()
}

func (w *worker) doRun(ctx context.Context) {
	for {
		select {
		case <-w.done:
			return
		case <-ctx.Done():
			return
		default:
		}

		start := time.Now()
		_, err := w.c.calcNthPrime(1000)
		end := time.Now()

		w.data.QueueClientRequest(w.run, &sqlite.AddRequestParams{
			WorkerID:  w.id,
			StartTime: start,
			EndTime:   end,
			Success:   err == nil,
		})

	}
}

func (w *worker) Stop() {
	close(w.done)
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

func makeRequest(c *http.Client) error {
	resp, err := c.Get("http://localhost:8080/meow")
	if err != nil {
		return errors.Wrap(err, "sending request failed")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "reading response body failed")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unexpected stauts - %d %s; body - %s", resp.StatusCode, resp.Status, string(body))
	}

	return nil
}
