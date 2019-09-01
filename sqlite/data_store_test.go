package sqlite

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func newTestDataStore(t *testing.T) *DataStore {
	ds, err := NewDataStore(":memory:")
	require.NoError(t, err)
	require.NotNil(t, ds)
	return ds
}

func TestDataStore(t *testing.T) {
	ctx := context.Background()

	ds := newTestDataStore(t)
	err := ds.CreateTables(ctx)
	require.NoError(t, err)
	ds.Start()

	runID := newID()
	now := time.Now().UTC()
	startTime := now
	err = ds.WriteRunStart(ctx, &AddRunParams{
		ID:         runID,
		StartTime:  startTime,
		NumWorkers: 2,
	})
	require.NoError(t, err)

	run := &Run{
		ID:        runID,
		StartTime: startTime,
	}
	addReqParams := &AddRequestParams{
		WorkerID:  2,
		StartTime: now,
		EndTime:   now.Add(time.Second),
		Success:   true,
	}
	ds.QueueClientRequest(run, addReqParams)

	addTCPConnParams := &AddTCPConnParams{
		Time:        now,
		Established: 2,
	}
	ds.QueueTCPConn(run, addTCPConnParams)

	endTime := startTime.Add(time.Minute)
	err = ds.WriteRunEnd(ctx, runID, endTime)
	require.NoError(t, err)

	err = ds.StopAndClose()
	require.NoError(t, err)
}
