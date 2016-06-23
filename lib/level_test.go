package ecslogs

import "testing"

var levelTests = []struct {
	lvl Level
	str string
}{
	{
		lvl: EMERGENCY,
		str: "EMERGENCY",
	},
	{
		lvl: ALERT,
		str: "ALERT",
	},
	{
		lvl: CRITICAL,
		str: "CRITICAL",
	},
	{
		lvl: ERROR,
		str: "ERROR",
	},
	{
		lvl: WARNING,
		str: "WARNING",
	},
	{
		lvl: NOTICE,
		str: "NOTICE",
	},
	{
		lvl: INFO,
		str: "INFO",
	},
	{
		lvl: DEBUG,
		str: "DEBUG",
	},
}

func TestParseLevelSuccess(t *testing.T) {
	for _, test := range levelTests {
		if lvl, err := ParseLevel(test.str); err != nil {
			t.Errorf("%s: error: %s", test.str, err)
		} else if lvl != test.lvl {
			t.Errorf("%s: invalid level: %s", test.str, lvl)
		}
	}
}

func TestParseLevelFailure(t *testing.T) {
	if _, err := ParseLevel(""); err == nil {
		t.Error("no error returned when parsing an invalid log level")
	} else if s := err.Error(); s != "invalid message level \"\"" {
		t.Error("invalid error message returned when parsing an invalid log level:", s)
	}
}

func TestLevelString(t *testing.T) {
	for _, test := range levelTests {
		if s := test.lvl.String(); s != test.str {
			t.Errorf("%s: invalid string: %s", test.lvl, s)
		}
	}
}
