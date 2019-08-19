package benchmark

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/pkg/errors"
)

type DataStore struct {
	db            *sql.DB
	requestParams chan *AddRequestParams
	lastID        int
}

func NewDataStore(filePath string) (*DataStore, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, errors.Wrap(err, "creating new data store failed")
	}

	requestParams := make(chan *AddRequestParams, 10)

	return &DataStore{
		db:            db,
		requestParams: requestParams,
	}, nil
}

func (d *DataStore) CreateTables() error {
	query := `
CREATE TABLE requests (
	id 			INTEGER 		PRIMARY KEY,
	worker_id 	INTEGER			NOT NULL,
	start_time 	DATETIME		NOT NULL,
	end_time 	DATETIME		NOT NULL,
	duration_ms	INTEGER			NOT NULL,
	success		INTEGER			NOT NULL,
);`
	_, err := d.db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "creating tables failed")
	}

	return nil
}

func (d *DataStore) Start() {

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

}

func (d *DataStore) writeRequests() {

	rs := make([]*AddRequestParams, 0, 1000)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case r := <-d.requestParams:
			rs = append(rs, r)
		case <-ticker.C:
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
		sqlParams := []interface{}{
			d.lastID,
			param.WorkerID,
			param.StartTime,
			param.EndTime,
			param.EndTime.Sub(param.StartTime) / time.Millisecond,
			param.Success,
		}
		_, err = tx.Exec(query, sqlParams)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()

}
