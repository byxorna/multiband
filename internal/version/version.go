package version

import (
	"fmt"
	"runtime/debug"

	"strconv"
	"time"
)

var (
	Commit            string
	Short             string
	RawBuildTimestamp string
)

func Verbose() string {
	v := Short
	if bi, ok := debug.ReadBuildInfo(); ok {
		v = bi.Main.Version
	}
	return fmt.Sprintf("%s (%s built %s)", v, Commit, BuiltAt().Format(time.RFC3339))
}

func BuiltAt() time.Time {
	ts, err := strconv.ParseInt(RawBuildTimestamp, 10, 64)
	if err != nil {
		return time.Time{}
	}
	t := time.Unix(ts, 0)

	return t
}
