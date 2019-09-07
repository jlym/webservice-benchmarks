package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/jlym/webservice-benchmarks/util"
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

	runID := util.NewID()
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
		RunID:       runID,
		Time:        now,
		Established: 2,
	}
	ds.QueueTCPConn(addTCPConnParams)

	addConnStatusParams := &AddConnStatusParams{
		RunID: runID,
		Time:  now,
		Fd:    23434,
	}
	ds.QueueConnStatus(addConnStatusParams)

	endTime := startTime.Add(time.Minute)
	err = ds.WriteRunEnd(ctx, runID, endTime)
	require.NoError(t, err)

	ds.Stop()

	tcpConns, err := getTCPConns(ctx, ds.db)
	require.NoError(t, err)
	require.Len(t, tcpConns, 1)
	tcpConn := tcpConns[0]
	require.NotNil(t, tcpConn)
	require.Equal(t, tcpConn.runID, runID)
	require.Equal(t, tcpConn.established, 2)

	clientRequests, err := getClientRequests(ctx, ds.db)
	require.NoError(t, err)
	require.Len(t, clientRequests, 1)
	clientRequest := clientRequests[0]
	require.NotNil(t, clientRequest)
	require.Equal(t, clientRequest.runID, runID)
	require.Equal(t, clientRequest.workerID, 2)

	connStatuses, err := getConnStatus(ctx, ds.db)
	require.NoError(t, err)
	require.Len(t, connStatuses, 1)
	connStatus := connStatuses[0]
	require.NotNil(t, connStatus)
	require.Equal(t, connStatus.runID, runID)
	require.Equal(t, connStatus.fd, uint32(23434))

	err = ds.Close()
	require.NoError(t, err)
}
