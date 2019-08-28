package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestRuns(t *testing.T) {
	ctx := context.Background()
	db := newInMemoryDb(t)
	defer db.Close()

	testInTransaction(t, db, func(ctx context.Context, tx *sql.Tx) error {
		return createRunsTable(ctx, tx)
	})

	startTime := time.Now().UTC()
	params := &AddRunParams{
		ID:         "runid",
		StartTime:  startTime,
		Desc:       "test desc",
		NumWorkers: 3,
	}

	err := insertIntoRuns(ctx, db, params)
	require.NoError(t, err)

	runs, err := getRuns(context.Background(), db)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	r := runs[0]

	require.Equal(t, params.ID, r.id)
	require.Equal(t, params.StartTime, r.startTime)
	require.NotNil(t, r.desc)
	require.Equal(t, params.Desc, *r.desc)
	require.NotNil(t, r.numWorkers)
	require.Equal(t, params.NumWorkers, *r.numWorkers)
	require.Nil(t, r.endTime)

	endTime := startTime.Add(time.Second * 30)
	err = updateRunEndTime(ctx, db, params.ID, endTime)
	require.NoError(t, err)

	runs, err = getRuns(context.Background(), db)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	r = runs[0]

	require.Equal(t, params.ID, r.id)
	require.NotNil(t, r.endTime)
	require.Equal(t, endTime, *r.endTime)
}
