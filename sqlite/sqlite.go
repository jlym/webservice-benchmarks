package sqlite

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

func newID() string {
	uuid := uuid.NewV4()
	return uuid.String()
}

type Run struct {
	ID        string
	StartTime time.Time
}

func (r *Run) secondsSinceStart(t time.Time) int {
	dur := t.Sub(r.StartTime)
	return int(dur / time.Second)
}

func (r *Run) millisecondsSinceStart(t time.Time) int {
	dur := t.Sub(r.StartTime)
	return int(dur / time.Millisecond)
}
