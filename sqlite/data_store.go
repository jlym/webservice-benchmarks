package sqlite

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/jlym/webservice-benchmarks/util"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type DataStore struct {
	db         *sql.DB
	writeQueue chan *writeQueueParams
	stopSender *util.StopSender
}

func NewDataStore(filePath string) (*DataStore, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, errors.Wrap(err, "creating new data store failed")
	}

	writeQueue := make(chan *writeQueueParams, 10000)

	return &DataStore{
		db:         db,
		writeQueue: writeQueue,
		stopSender: util.NewStopSender(),
	}, nil
}

func (d *DataStore) CreateTables(ctx context.Context) error {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "db - starting transaction failed")
	}

	err = createRunsTable(ctx, tx)
	err = rollbackTransaction(tx, err)
	if err != nil {
		return err
	}

	err = createClientRequestsTable(ctx, tx)
	err = rollbackTransaction(tx, err)
	if err != nil {
		return err
	}

	err = createTCPConnsTable(ctx, tx)
	err = rollbackTransaction(tx, err)
	if err != nil {
		return err
	}

	err = createConnStatusTable(ctx, tx)
	err = rollbackTransaction(tx, err)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "create tables - commit transaction failed")
	}

	return nil
}

func (d *DataStore) Start() {
	stopReciever := d.stopSender.NewReciever()
	go d.writeFromQueue(stopReciever)
}

func (d *DataStore) Stop() {
	d.stopSender.StopAndWait()
}

func (d *DataStore) Close() error {
	err := d.db.Close()
	if err != nil {
		return errors.Wrap(err, "closing db conn failed")
	}
	return nil
}

func (d *DataStore) WriteRunStart(ctx context.Context, params *AddRunParams) error {
	err := insertIntoRuns(ctx, d.db, params)
	if err != nil {
		return errors.Wrap(err, "write run start failed")
	}
	return nil
}

func (d *DataStore) WriteRunEnd(ctx context.Context, runID string, endTime time.Time) error {
	err := updateRunEndTime(ctx, d.db, runID, endTime)
	if err != nil {
		return errors.Wrap(err, "write run end failed")
	}
	return nil
}

func (d *DataStore) QueueTCPConn(run *Run, params *AddTCPConnParams) {
	if params == nil {
		return
	}

	d.writeQueue <- &writeQueueParams{
		run:     run,
		tcpConn: params,
	}
}

func (d *DataStore) QueueClientRequest(run *Run, params *AddRequestParams) {
	if params == nil {
		return
	}

	d.writeQueue <- &writeQueueParams{
		run:           run,
		clientRequest: params,
	}
}

func (d *DataStore) QueueConnStatus(run *Run, params *AddConnStatusParams) {
	if params == nil {
		return
	}

	d.writeQueue <- &writeQueueParams{
		run:        run,
		connStatus: params,
	}
}

type writeQueueParams struct {
	clientRequest *AddRequestParams
	tcpConn       *AddTCPConnParams
	connStatus    *AddConnStatusParams
	run           *Run
}

func (d *DataStore) writeFromQueue(stopReiever *util.StopReciever) {
	defer stopReiever.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for stopReiever.ShouldContinue() {
		buffer := make([]*writeQueueParams, 0, 1000)

		hasMore := true
		for hasMore {
			select {
			case params := <-d.writeQueue:
				buffer = append(buffer, params)
			case <-ticker.C:
				hasMore = false
			case <-stopReiever.ShouldStopC:
				hasMore = false
			}
		}

		err := d.writeToDB(buffer)
		if err != nil {
			log.Println(err)
		}
	}

	buffer := make([]*writeQueueParams, 0, 1000)
	hasMore := true
	for hasMore {
		select {
		case params := <-d.writeQueue:
			buffer = append(buffer, params)
		default:
			hasMore = false
		}
	}

	err := d.writeToDB(buffer)
	if err != nil {
		log.Println(err)
	}
}

func (d *DataStore) writeToDB(params []*writeQueueParams) error {
	ctx := context.Background()

	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "db - starting transaction failed")
	}

	for _, param := range params {
		if param.clientRequest != nil {
			err = insertIntoClientRequests(ctx, tx, param.run, param.clientRequest)
			err = rollbackTransaction(tx, err)
			if err != nil {
				return err
			}
		}

		if param.tcpConn != nil {
			err = insertIntoTCPConns(ctx, tx, param.run, param.tcpConn)
			err = rollbackTransaction(tx, err)
			if err != nil {
				return err
			}
		}

		if param.connStatus != nil {
			err = insertIntoConnStatus(ctx, tx, param.run, param.connStatus)
			err = rollbackTransaction(tx, err)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func rollbackTransaction(tx *sql.Tx, err error) error {
	if err == nil {
		return nil
	}

	rollbackErr := tx.Rollback()
	if rollbackErr != nil {
		return rollbackErr
	}

	return err
}
