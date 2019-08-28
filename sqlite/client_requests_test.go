package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestClientRequests(t *testing.T) {
	db := newInMemoryDb(t)
	defer db.Close()

	testInTransaction(t, db, func(ctx context.Context, tx *sql.Tx) error {
		return createClientRequestsTable(ctx, tx)
	})

	params := &AddRequestParams{
		WorkerID:  1,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Success:   true,
		Error:     "err",
	}
	run := &Run{
		ID:        "runid",
		StartTime: time.Now(),
	}
	testInTransaction(t, db, func(ctx context.Context, tx *sql.Tx) error {
		return insertIntoClientRequests(ctx, tx, run, params)
	})

	clientRequests, err := getClientRequests(context.Background(), db)
	require.NoError(t, err)
	require.Len(t, clientRequests, 1)
	c := clientRequests[0]

	require.Equal(t, params.WorkerID, c.workerID)
}

func newInMemoryDb(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, db)
	return db
}

func testInTransaction(t *testing.T, db *sql.DB, f func(ctx context.Context, tx *sql.Tx) error) {
	tx, err := db.Begin()
	require.NoError(t, err)

	err = f(context.Background(), tx)
	if err != nil {
		tx.Rollback()
		require.NoError(t, err)
	}

	err = tx.Commit()
	require.NoError(t, err)
}
