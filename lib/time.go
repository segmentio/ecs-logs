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
		if t, err = time.Parse(format, s); err == nil {
			ts = MakeTimestamp(t)
			break
		}
	}

	return
}

func MakeTimestamp(t time.Time) Timestamp {
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

func (ts Timestamp) Time() time.Time {
	return time.Unix(ts.Seconds(), ts.Nanoseconds()%1000000000)
}

func (ts Timestamp) Format(format string) string {
	return ts.Time().Format(format)
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
