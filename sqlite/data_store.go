package sqlite

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type DataStore struct {
	db           *sql.DB
	writeQueue   chan *writeQueueParams
	done         chan struct{}
	bgThreadDone chan struct{}
}

func NewDataStore(filePath string) (*DataStore, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, errors.Wrap(err, "creating new data store failed")
	}

	writeQueue := make(chan *writeQueueParams, 10000)
	done := make(chan struct{})
	bgThreadDone := make(chan struct{})

	return &DataStore{
		db:           db,
		writeQueue:   writeQueue,
		done:         done,
		bgThreadDone: bgThreadDone,
	}, nil
}

func newDataStoreWithInMemoryDb() (*DataStore, error) {
	return NewDataStore(":memory:")
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

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "create tables - commit transaction failed")
	}

	return nil
}

func (d *DataStore) Start() {
	go d.writeRequests()
}

func (d *DataStore) StopAndClose() error {
	close(d.done)
	<-d.bgThreadDone

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

type writeQueueParams struct {
	clientRequest *AddRequestParams
	tcpConn       *AddTCPConnParams
	run           *Run
}

func (d *DataStore) writeRequests() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for !d.isDone() {
		rs := make([]*writeQueueParams, 0, 1000)

		shouldCollectRequests := true
		for shouldCollectRequests {
			select {
			case r := <-d.writeQueue:
				rs = append(rs, r)
			case <-ticker.C:
				shouldCollectRequests = false
			case <-d.done:
				shouldCollectRequests = false
			}
		}

		err := d.writeRequestsToDB(rs)
		if err != nil {
			log.Println(err)
		}
	}

	close(d.bgThreadDone)
}

func (d *DataStore) writeRequestsToDB(requestParams []*writeQueueParams) error {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "db - starting transaction failed")
	}

	ctx := context.Background()

	for _, param := range requestParams {
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
	}

	return tx.Commit()
}

func (d *DataStore) isDone() bool {
	select {
	case <-d.done:
		return true
	default:
		return false
	}
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
