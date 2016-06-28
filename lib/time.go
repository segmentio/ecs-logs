package ecslogs

import (
	"strconv"
	"time"
)

type Timestamp int64

func Now() Timestamp {
	return MakeTimestamp(time.Now())
}

func ParseTimestamp(s string) (ts Timestamp, err error) {
	var t time.Time

	for _, format := range timeFormats {
		if t, err = time.ParseInLocation(format, s, time.UTC); err == nil {
			ts = MakeTimestamp(t)
			break
		}
	}

	return
}

func MakeTimestamp(t time.Time) Timestamp {
	t = t.UTC()
	return Timestamp(t.Unix()*1000000) + Timestamp(t.Nanosecond()/1000)
}

func (ts Timestamp) GoString() string {
	return "Timestamp(" + strconv.FormatInt(int64(ts), 10) + ")"
}

func (ts Timestamp) String() string {
	return ts.Format("2006-01-02T15:04:05.999999Z07:00")
}

func (ts Timestamp) Seconds() int64 {
	return int64(ts / 1000000)
}

func (ts Timestamp) Milliseconds() int64 {
	return int64(ts / 1000)
}

func (ts Timestamp) Microseconds() int64 {
	return int64(ts)
}

func (ts Timestamp) Nanoseconds() int64 {
	return int64(ts * 1000)
}

func (ts Timestamp) Time(tz *time.Location) time.Time {
	t1 := time.Unix(ts.Seconds(), ts.Nanoseconds()%1000000000)
	t2 := t1.In(tz)
	return t2.Add(t1.Sub(t2))
}

func (ts Timestamp) Format(format string) string {
	return ts.Time(time.UTC).Format(format)
}

func (ts Timestamp) MarshalText() (b []byte, err error) {
	b = []byte(ts.String())
	return
}

func (ts *Timestamp) UnmarshalText(b []byte) (err error) {
	*ts, err = ParseTimestamp(string(b))
	return
}

var (
	timeFormats = []string{
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999999Z07:00", // RFC3338 micro
		"2006-01-02T15:04:05.999Z07:00",    // RFC3338 milli
		time.RFC3339,
	}
)

const (
	secondsPerMinute       = 60
	secondsPerHour         = 60 * 60
	secondsPerDay          = 24 * secondsPerHour
	unixToInternal   int64 = (1969*365 + 1969/4 - 1969/100 + 1969/400) * secondsPerDay
)
