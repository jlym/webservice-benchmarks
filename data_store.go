package benchmark

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type DataStore struct {
	db            *sql.DB
	requestParams chan *AddRequestParams
	done          chan struct{}
}

func NewDataStore(filePath string) (*DataStore, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, errors.Wrap(err, "creating new data store failed")
	}

	requestParams := make(chan *AddRequestParams, 10)
	done := make(chan struct{})

	return &DataStore{
		db:            db,
		requestParams: requestParams,
		done:          done,
	}, nil
}

func newDataStoreWithInMemoryDb() (*DataStore, error) {
	return NewDataStore(":memory:")
}

func (d *DataStore) CreateTables(ctx context.Context) error {
	query := `
CREATE TABLE IF NOT EXISTS requests (
	id 			TEXT 		PRIMARY KEY,
	worker_id 	INTEGER			NOT NULL,
	start_time 	DATETIME		NOT NULL,
	end_time 	DATETIME		NOT NULL,
	duration_ms	INTEGER			NOT NULL,
	success		INTEGER			NOT NULL
);`
	_, err := d.db.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "creating tables failed")
	}

	return nil
}

func (d *DataStore) Start() {
	go d.writeRequests()
}

func (d *DataStore) Stop() {
	close(d.done)
}

type AddRequestParams struct {
	WorkerID  int
	StartTime time.Time
	EndTime   time.Time
	Success   bool
}

func (d *DataStore) AddRequestAsync(params *AddRequestParams) {
	if params == nil {
		return
	}

	d.requestParams <- params
}

func (d *DataStore) isDone() bool {
	select {
	case <-d.done:
		return true
	default:
		return false
	}
}

func (d *DataStore) writeRequests() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for !d.isDone() {
		rs := make([]*AddRequestParams, 0, 1000)

		shouldCollectRequests := true
		for shouldCollectRequests {
			select {
			case r := <-d.requestParams:
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
}

func (d *DataStore) writeRequestsToDB(requestParams []*AddRequestParams) error {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "db - starting transaction failed")
	}

	for _, param := range requestParams {
		query := `
INSERT INTO requests (id, worker_id, start_time, end_time, duration_ms, success)		
VALUES ($1, $2, $3, $4, $5, $6);
`

		uuid := uuid.NewV4()
		id := uuid.String()
		sqlParams := []interface{}{
			id,
			param.WorkerID,
			param.StartTime,
			param.EndTime,
			param.EndTime.Sub(param.StartTime) / time.Millisecond,
			param.Success,
		}
		_, err = tx.Exec(query, sqlParams...)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

type requestInfo struct {
	id         string
	workerID   int
	startTime  time.Time
	endTime    time.Time
	durationMs int
	success    bool
}

func (d *DataStore) getRequests(ctx context.Context) ([]*requestInfo, error) {
	query := `
SELECT id, worker_id, start_time, end_time, duration_ms, success
FROM requests;
`
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "db - get requests failed")
	}
	defer rows.Close()

	results := make([]*requestInfo, 0)
	for rows.Next() {
		var id string
		var workerID int
		var startTime time.Time
		var endTime time.Time
		var durationMs int
		var success bool

		if err := rows.Scan(&id, &workerID, &startTime, &endTime, &durationMs, &success); err != nil {
			return nil, errors.Wrap(err, "db - getting requests - scanning failed")
		}

		results = append(results, &requestInfo{
			id:         id,
			workerID:   workerID,
			startTime:  startTime,
			endTime:    endTime,
			durationMs: durationMs,
			success:    success,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "db - getting requests - scaning failed")
	}

	return results, nil

}
