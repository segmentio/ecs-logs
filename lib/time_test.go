package ecslogs

import "testing"

func TestParseTimestampSuccess(t *testing.T) {
	tests := []struct {
		t Timestamp
		s string
	}{
		{
			t: Timestamp(1467136754000000),
			s: "2016-06-28T17:59:14Z",
		},
	}

	for _, test := range tests {
		if ts, err := ParseTimestamp(test.s); err != nil {
			t.Errorf("%s: %s", test.s, err)
		} else if ts != test.t {
			t.Errorf("%s: %s %#v %#v", test.s, ts, ts, test.t)
		}
	}
}
