package framer

import (
	"time"
)

func getTime(timeInSeconds int64) time.Time {
	return time.Unix(timeInSeconds, 0)
}
