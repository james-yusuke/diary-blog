package site

import "time"

func timeFromUnixMilliseconds(milliseconds float64) time.Time {
	return time.UnixMilli(int64(milliseconds)).UTC()
}
