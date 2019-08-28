package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

type AddRequestParams struct {
	WorkerID  int
	StartTime time.Time
	EndTime   time.Time
	Success   bool
	Error     string
}

func createClientRequestsTable(ctx context.Context, tx *sql.Tx) error {
	query := `
	CREATE TABLE IF NOT EXISTS client_requests (
		id 				TEXT 		PRIMARY KEY,
		run_id			TEXT		NOT NULL,
		worker_id 		INTEGER		NOT NULL,
		start_time		DATETIME	NOT NULL,
		end_time		DATETIME	NOT NULL,
	
		s_since_start 	INTEGER		NOT NULL,
		ms_since_start	INTEGER 	NOT NULL,
	
		duration_ms		INTEGER		NOT NULL,
		success			INTEGER		NOT NULL,
		error			TEXT		NOT NULL
	);`

	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "creating client_requests table failed")
	}

	return nil
}

func insertIntoClientRequests(ctx context.Context, tx *sql.Tx, run *Run, params *AddRequestParams) error {
	query := `
		INSERT INTO client_requests (
			id, run_id, worker_id, start_time, end_time, s_since_start, ms_since_start,
			duration_ms, success, error)		
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		);`

	args := []interface{}{
		newID(),
		run.ID,
		params.WorkerID,
		params.StartTime,
		params.EndTime,
		run.secondsSinceStart(params.StartTime),
		run.millisecondsSinceStart(params.StartTime),
		params.EndTime.Sub(params.StartTime) / time.Millisecond,
		params.Success,
		params.Error,
	}

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "insert into client_requests failed")
	}

	return nil
}

type clientRequest struct {
	id                     string
	runID                  string
	workerID               int
	startTime              time.Time
	endTime                time.Time
	secondsSinceStart      int
	millisecondsSinceStart int
	durationMs             int
	success                bool
	errMessage             string
}

func getClientRequests(ctx context.Context, db *sql.DB) ([]*clientRequest, error) {
	query := `
		SELECT 
			id, run_id, worker_id, start_time, end_time, s_since_start, 
			ms_since_start, duration_ms, success, error
		FROM client_requests;`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "db - get client requests failed")
	}
	defer rows.Close()

	results := make([]*clientRequest, 0)
	for rows.Next() {
		r := clientRequest{}

		err := rows.Scan(
			&r.id,
			&r.runID,
			&r.workerID,
			&r.startTime,
			&r.endTime,
			&r.secondsSinceStart,
			&r.millisecondsSinceStart,
			&r.durationMs,
			&r.success,
			&r.errMessage)
		if err != nil {
			return nil, errors.Wrap(err, "db - getting client requests - scanning failed")
		}

		results = append(results, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "db - getting client requests - scaning failed")
	}

	return results, nil
}
