package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/jlym/webservice-benchmarks/util"
	"github.com/pkg/errors"
)

type AddConnStatusParams struct {
	RunID string
	Time  time.Time

	Fd          uint32
	Type        string
	LocalIP     string
	LocalPort   uint32
	RemoteIP    string
	RemotePort  uint32
	Status      string
	ProcessID   uint32
	ProcessName string
}

func createConnStatusTable(ctx context.Context, tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS conn_status (
			id 				TEXT 		PRIMARY KEY,
			run_id 			TEXT 		NOT NULL,
			time 			DATETIME 	NOT NULL,

			fd 				INTEGER 	NOT NULL,
			type			TEXT 		NOT NULL,
			local_ip 		TEXT 		NOT NULL,
			local_port 		INTEGER 	NOT NULL,
			remote_ip 		TEXT 		NOT NULL,
			remote_port 	INTEGER 	NOT NULL,
			status 			TEXT 		NOT NULL,
			process_id		INTEGER		NOT NULL,
			process_name	TEXT		NOT NULL
		);`

	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "creating conn_status table failed")
	}

	return nil
}

func insertIntoConnStatus(ctx context.Context, tx *sql.Tx, params *AddConnStatusParams) error {
	query := `
		INSERT INTO conn_status (
			id, run_id, time, 
			fd, type, local_ip, local_port, remote_ip, remote_port, status,
			process_id, process_name)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`

	args := []interface{}{
		util.NewID(),
		params.RunID,
		params.Time,

		params.Fd,
		params.Type,
		params.LocalIP,
		params.LocalPort,
		params.RemoteIP,
		params.RemotePort,
		params.Status,
		params.ProcessID,
		params.ProcessName,
	}

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "insert into conn_status failed")
	}

	return nil
}

type connStatus struct {
	id    string
	runID string
	time  time.Time

	fd          uint32
	connType    string
	localIP     string
	localPort   uint32
	remoteIP    string
	remotePort  uint32
	status      string
	processID   uint32
	processName string
}

func getConnStatus(ctx context.Context, db *sql.DB) ([]*connStatus, error) {
	query := `
		SELECT 
			id, run_id, time, 
			fd, type, local_ip, local_port, remote_ip, remote_port, status,
			process_id, process_name
		FROM conn_status;`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "db - get conn status failed")
	}
	defer rows.Close()

	results := make([]*connStatus, 0)
	for rows.Next() {
		r := connStatus{}

		err := rows.Scan(
			&r.id,
			&r.runID,
			&r.time,
			&r.fd,
			&r.connType,
			&r.localIP,
			&r.localPort,
			&r.remoteIP,
			&r.remotePort,
			&r.status,
			&r.processID,
			&r.processName,
		)
		if err != nil {
			return nil, errors.Wrap(err, "db - getting conn status - scanning failed")
		}

		results = append(results, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "db - getting conn status - scaning failed")
	}

	return results, nil
}
