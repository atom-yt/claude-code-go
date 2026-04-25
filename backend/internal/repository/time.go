package repository

import (
	"time"
)

// timeNow returns the current UTC time
func timeNow() time.Time {
	return time.Now().UTC()
}