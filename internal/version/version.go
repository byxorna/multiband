package version

import (
	"strconv"
	"time"
)

var (
	Commit            string
	Build             string
	RawBuildTimestamp string
)

func BuiltAt() time.Time {
	ts, err := strconv.ParseInt(RawBuildTimestamp, 10, 64)
	if err != nil {
		return time.Time{}
	}
	t := time.Unix(ts, 0)

	return t
}
