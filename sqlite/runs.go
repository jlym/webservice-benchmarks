package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

type AddRunParams struct {
	ID         string
	StartTime  time.Time
	Desc       string
	NumWorkers int
}

func createRunsTable(ctx context.Context, tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS runs (
			id 				TEXT 		PRIMARY KEY,
			start_time 		DATETIME 	NOT NULL,
			end_time 		DATETIME,
			desc 			TEXT,
			num_workers 	INTEGER
		);`

	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "creating runs table failed")
	}

	return nil
}

func insertIntoRuns(ctx context.Context, db *sql.DB, params *AddRunParams) error {
	query := `
		INSERT INTO runs (id, start_time, desc, num_workers)		
		VALUES ($1, $2, $3, $4);`

	args := []interface{}{
		params.ID,
		params.StartTime,
		params.Desc,
		params.NumWorkers,
	}

	_, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "insert into runs failed")
	}

	return nil
}

func updateRunEndTime(ctx context.Context, db *sql.DB, runID string, endTime time.Time) error {
	query := `
		UPDATE runs
		SET end_time = $1
		WHERE id = $2;`
	args := []interface{}{endTime, runID}

	_, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "update end time run failed")
	}

	return nil
}

type run struct {
	id         string
	startTime  time.Time
	endTime    *time.Time
	desc       *string
	numWorkers *int
}

func getRuns(ctx context.Context, db *sql.DB) ([]*run, error) {
	query := `
		SELECT id, start_time, end_time, desc, num_workers
		FROM runs;`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "db - get runs failed")
	}
	defer rows.Close()

	results := make([]*run, 0)
	for rows.Next() {
		r := run{}

		err := rows.Scan(
			&r.id,
			&r.startTime,
			&r.endTime,
			&r.desc,
			&r.numWorkers)
		if err != nil {
			return nil, errors.Wrap(err, "db - get runs failed - scanning failed")
		}

		results = append(results, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "db - get runs failed - scaning failed")
	}

	return results, nil
}
