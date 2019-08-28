package sqlite

/*
func TestWritingRequestInfo(t *testing.T) {
	db, err := newDataStoreWithInMemoryDb()
	require.NoError(t, err)
	require.NotNil(t, db)

	err = db.CreateTables(context.Background())
	require.NoError(t, err)

	req1 := AddRequestParams{
		WorkerID:  1,
		StartTime: time.Date(2019, time.February, 1, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2019, time.February, 1, 1, 0, 0, 0, time.UTC),
		Success:   true,
	}
	req2 := AddRequestParams{
		WorkerID:  2,
		StartTime: time.Date(2019, time.February, 1, 4, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2019, time.February, 1, 5, 0, 0, 0, time.UTC),
		Success:   false,
	}

	err = db.writeRequestsToDB([]*AddRequestParams{
		&req1, &req2,
	})
	require.NoError(t, err)

	reqs, err := db.getRequests(context.Background())
	require.NoError(t, err)
	require.Len(t, reqs, 2)

	require.NotEmpty(t, reqs[0].id)
	require.Equal(t, reqs[0].startTime, req1.StartTime)
	require.Equal(t, reqs[0].endTime, req1.EndTime)
	require.Equal(t, reqs[0].workerID, req1.WorkerID)
	require.Equal(t, reqs[0].success, req1.Success)

	require.NotEmpty(t, reqs[1].id)
	require.Equal(t, reqs[1].startTime, req2.StartTime)
	require.Equal(t, reqs[1].endTime, req2.EndTime)
	require.Equal(t, reqs[1].workerID, req2.WorkerID)
	require.Equal(t, reqs[1].success, req2.Success)
}


func TestAddRequestAsync(t *testing.T) {
	db, err := newDataStoreWithInMemoryDb()
	require.NoError(t, err)
	require.NotNil(t, db)

	err = db.CreateTables(context.Background())
	require.NoError(t, err)

	db.Start()

	req1 := &AddRequestParams{
		WorkerID:  1,
		StartTime: time.Date(2019, time.February, 1, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2019, time.February, 1, 1, 0, 0, 0, time.UTC),
		Success:   true,
	}
	req2 := &AddRequestParams{
		WorkerID:  2,
		StartTime: time.Date(2019, time.February, 1, 4, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2019, time.February, 1, 5, 0, 0, 0, time.UTC),
		Success:   false,
	}

	db.QueueClientRequest(req1)
	time.Sleep(time.Second * 2)

	reqs, err := db.getRequests(context.Background())
	require.Len(t, reqs, 1)

	require.NotEmpty(t, reqs[0].id)
	require.Equal(t, reqs[0].startTime, req1.StartTime)
	require.Equal(t, reqs[0].endTime, req1.EndTime)
	require.Equal(t, reqs[0].workerID, req1.WorkerID)
	require.Equal(t, reqs[0].success, req1.Success)

	db.QueueClientRequest(req2)
	time.Sleep(time.Second * 2)

	reqs, err = db.getRequests(context.Background())
	require.Len(t, reqs, 2)

	require.NotEmpty(t, reqs[0].id)
	require.Equal(t, reqs[0].startTime, req1.StartTime)
	require.Equal(t, reqs[0].endTime, req1.EndTime)
	require.Equal(t, reqs[0].workerID, req1.WorkerID)
	require.Equal(t, reqs[0].success, req1.Success)

	require.NotEmpty(t, reqs[1].id)
	require.Equal(t, reqs[1].startTime, req2.StartTime)
	require.Equal(t, reqs[1].endTime, req2.EndTime)
	require.Equal(t, reqs[1].workerID, req2.WorkerID)
	require.Equal(t, reqs[1].success, req2.Success)

	db.Stop()
}
*/
