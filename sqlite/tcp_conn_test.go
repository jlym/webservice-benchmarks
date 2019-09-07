package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestTCPConns(t *testing.T) {
	db := newInMemoryDb(t)
	defer db.Close()

	testInTransaction(t, db, func(ctx context.Context, tx *sql.Tx) error {
		return createTCPConnsTable(ctx, tx)
	})

	params := &AddTCPConnParams{
		Time:        time.Now().UTC(),
		RunID:       "runid",
		Established: 1,
		SynSent:     2,
		SynRecv:     3,
		FinWait1:    4,
		FinWait2:    5,
		TimeWait:    6,
		Close:       7,
		CloseWait:   8,
		LastAck:     9,
		Listen:      10,
		Closing:     11,
	}
	testInTransaction(t, db, func(ctx context.Context, tx *sql.Tx) error {
		return insertIntoTCPConns(ctx, tx, params)
	})

	tcpConns, err := getTCPConns(context.Background(), db)
	require.NoError(t, err)
	require.Len(t, tcpConns, 1)
	c := tcpConns[0]

	require.Equal(t, params.Time, c.time)
	require.Equal(t, params.RunID, c.runID)
	require.Equal(t, params.Established, c.established)
	require.Equal(t, params.SynSent, c.synSent)
	require.Equal(t, params.SynRecv, c.synRecv)
	require.Equal(t, params.TimeWait, c.timeWait)
	require.Equal(t, params.Close, c.close)
	require.Equal(t, params.CloseWait, c.closeWait)
	require.Equal(t, params.LastAck, c.lastAck)
	require.Equal(t, params.Listen, c.listen)
	require.Equal(t, params.Closing, c.closing)
}
