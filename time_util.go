package throughcache

import (
	"fmt"
	"time"
)

func ElapsedMilliseconds(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds() / 1000.0 / 1000.0)
}

func ElapsedMillisecondsStr(start time.Time) string {
	elapsed := ElapsedMilliseconds(start)
	return fmt.Sprintf("%.2f ms", elapsed)
}

func ElapsedMicrosecond(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds() / 1000.0)
}

func ElapsedSeconds(start time.Time) float64 {
	return time.Since(start).Seconds()
}
