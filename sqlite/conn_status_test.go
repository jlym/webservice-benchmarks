package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestConnStatus(t *testing.T) {
	db := newInMemoryDb(t)
	defer db.Close()

	testInTransaction(t, db, func(ctx context.Context, tx *sql.Tx) error {
		return createConnStatusTable(ctx, tx)
	})

	params := &AddConnStatusParams{
		Time:        time.Now().UTC(),
		Fd:          1,
		Type:        "tcp",
		LocalIP:     "10.0.0.1",
		LocalPort:   64232,
		RemoteIP:    "62.4.32.1",
		RemotePort:  80,
		Status:      "ESTABLISHED",
		ProcessID:   3332,
		ProcessName: "test.exe",
	}
	run := &Run{
		ID:        "runid",
		StartTime: time.Now(),
	}
	testInTransaction(t, db, func(ctx context.Context, tx *sql.Tx) error {
		return insertIntoConnStatus(ctx, tx, run, params)
	})

	connStatuses, err := getConnStatus(context.Background(), db)
	require.NoError(t, err)
	require.Len(t, connStatuses, 1)
	c := connStatuses[0]

	require.Equal(t, params.Time, c.time)
	require.Equal(t, run.ID, c.runID)
	require.Equal(t, params.Fd, c.fd)
	require.Equal(t, params.Type, c.connType)
	require.Equal(t, params.LocalIP, c.localIP)
	require.Equal(t, params.LocalPort, c.localPort)
	require.Equal(t, params.RemoteIP, c.remoteIP)
	require.Equal(t, params.RemotePort, c.remotePort)
	require.Equal(t, params.Status, c.status)
	require.Equal(t, params.ProcessID, c.processID)
	require.Equal(t, params.ProcessName, c.processName)
}
