package common

import "time"

var layout string = "15:04:05.000000"

func FormatTime(t time.Time) string {
	return t.Format(layout)
}

func ParseTime(timeStr string) (time.Time, error) {
	res, err := time.Parse(layout, timeStr)
	if err == nil {
		return res, err
	}

	// backward compatible
	return time.Parse(time.RFC3339Nano, timeStr)

}
