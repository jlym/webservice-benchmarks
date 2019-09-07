package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/jlym/webservice-benchmarks/util"
	"github.com/pkg/errors"
)

type AddTCPConnParams struct {
	RunID string
	Time  time.Time

	Established int
	SynSent     int
	SynRecv     int
	FinWait1    int
	FinWait2    int
	TimeWait    int
	Close       int
	CloseWait   int
	LastAck     int
	Listen      int
	Closing     int
}

func createTCPConnsTable(ctx context.Context, tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS tcp_conns (
			id 				TEXT 		PRIMARY KEY,
			run_id 			TEXT 		NOT NULL,
			time 			DATETIME 	NOT NULL,

			established 	INTEGER 	NOT NULL,
			syn_sent		INTEGER		NOT NULL,
			syn_recv		INTEGER		NOT NULL,
			fin_wait_1 		INTEGER 	NOT NULL,
			fin_wait_2 		INTEGER 	NOT NULL,
			time_wait 		INTEGER 	NOT NULL,
			close			INTEGER		NOT NULL,
			close_wait 		INTEGER 	NOT NULL,
			last_ack 		INTEGER 	NOT NULL,
			listen			INTEGER		NOT NULL,
			closing			INTEGER		NOT NULL
		);`

	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "creating tcp_conns table failed")
	}

	return nil
}

func insertIntoTCPConns(ctx context.Context, tx *sql.Tx, params *AddTCPConnParams) error {
	query := `
		INSERT INTO tcp_conns (
			id, run_id, time, established, syn_sent, 
			syn_recv, fin_wait_1, fin_wait_2, time_wait, close, close_wait,
			last_ack, listen, closing)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);`

	args := []interface{}{
		util.NewID(),
		params.RunID,
		params.Time,

		params.Established,
		params.SynSent,
		params.SynRecv,
		params.FinWait1,
		params.FinWait2,
		params.TimeWait,
		params.Close,
		params.CloseWait,
		params.LastAck,
		params.Listen,
		params.Closing,
	}

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "insert into tcp_conns failed")
	}

	return nil
}

type tcpConn struct {
	id          string
	runID       string
	time        time.Time
	established int
	synSent     int
	synRecv     int
	finWait1    int
	finWait2    int
	timeWait    int
	close       int
	closeWait   int
	lastAck     int
	listen      int
	closing     int
}

func getTCPConns(ctx context.Context, db *sql.DB) ([]*tcpConn, error) {
	query := `
		SELECT 
			id, run_id, time, established, syn_sent, 
			syn_recv, fin_wait_1, fin_wait_2, time_wait, close, close_wait,
			last_ack, listen, closing
		FROM tcp_conns;`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "db - get tcp conns failed")
	}
	defer rows.Close()

	results := make([]*tcpConn, 0)
	for rows.Next() {
		r := tcpConn{}

		err := rows.Scan(
			&r.id,
			&r.runID,
			&r.time,
			&r.established,
			&r.synSent,
			&r.synRecv,
			&r.finWait1,
			&r.finWait2,
			&r.timeWait,
			&r.close,
			&r.closeWait,
			&r.lastAck,
			&r.listen,
			&r.closing,
		)
		if err != nil {
			return nil, errors.Wrap(err, "db - getting tcp conns - scanning failed")
		}

		results = append(results, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "db - getting tcp conns - scaning failed")
	}

	return results, nil
}
