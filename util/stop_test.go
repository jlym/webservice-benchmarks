package util

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestStop(t *testing.T) {
	stopSender := NewStopSender()
	require.NotNil(t, stopSender)

	stopReciever := stopSender.NewReciever()
	require.NotNil(t, stopSender)

	shouldStop := stopReciever.ShouldStop()
	require.False(t, shouldStop)

	shouldContinue := stopReciever.ShouldContinue()
	require.True(t, shouldContinue)

	ch := make(chan struct{})
	go func() {
		stopSender.StopAndWait()
		close(ch)
	}()
	go func() {
		<-stopReciever.ShouldStopC
		stopReciever.Done()
	}()

	<-ch

	shouldStop = stopReciever.ShouldStop()
	require.True(t, shouldStop)

	shouldContinue = stopReciever.ShouldContinue()
	require.False(t, shouldContinue)
}
