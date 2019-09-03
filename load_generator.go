package webservice_benchmarks

import (
	"context"
	"log"
	"time"

	"github.com/jlym/webservice-benchmarks/sqlite"
	"github.com/jlym/webservice-benchmarks/util"
)

type SendRequestFunc func(workerID int) error

type TestConfig struct {
	DBFilePath     string
	NumWorkers     int
	RampUpDuration time.Duration
	TestDuration   time.Duration
	Port           int
	RunID          string
}

func GenerateLoad(config *TestConfig, f SendRequestFunc) error {
	ctx := context.Background()

	data, err := sqlite.NewDataStore(config.DBFilePath)
	if err != nil {
		return err
	}

	err = data.CreateTables(ctx)
	if err != nil {
		return err
	}

	data.Start()
	defer func() {
		data.Stop()
		_ = data.Close()
	}()

	run := &sqlite.Run{
		ID:        config.RunID,
		StartTime: time.Now().UTC(),
	}
	err = data.WriteRunStart(ctx, &sqlite.AddRunParams{
		ID:         run.ID,
		StartTime:  run.StartTime,
		NumWorkers: config.NumWorkers,
	})
	if err != nil {
		return err
	}

	stopSender := util.NewStopSender()

	for workerID := 0; workerID < config.NumWorkers; workerID++ {
		log.Println("start: ", workerID)
		go doActionRepeatedly(
			stopSender.NewReciever(),
			data,
			run,
			workerID,
			f)

		if workerID < config.NumWorkers-1 {
			ch := time.After(config.RampUpDuration)
			<-ch
		}
	}

	ch := time.After(config.TestDuration)
	<-ch

	stopSender.StopAndWait()
	return nil
}

func doActionRepeatedly(
	stopReciever *util.StopReciever,
	db *sqlite.DataStore,
	run *sqlite.Run,
	workerID int,
	f SendRequestFunc) {

	defer stopReciever.Done()

	for stopReciever.ShouldContinue() {
		start := time.Now().UTC()
		err := f(workerID)
		end := time.Now().UTC()

		errorMessage := ""
		if err != nil {
			errorMessage = err.Error()
		}

		db.QueueClientRequest(run, &sqlite.AddRequestParams{
			WorkerID:  workerID,
			StartTime: start,
			EndTime:   end,
			Success:   err == nil,
			Error:     errorMessage,
		})
	}
}
