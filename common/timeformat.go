package common

import "time"

var layout string = "15:04:05.000000"

func FormatTime(t time.Time) string {
	return t.Format(layout)
}

func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse(layout, timeStr)
}
