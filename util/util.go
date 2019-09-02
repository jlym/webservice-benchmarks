package util

import (
	uuid "github.com/satori/go.uuid"
)

// NewID returns a v4 UUID.
func NewID() string {
	uuid := uuid.NewV4()
	return uuid.String()
}
