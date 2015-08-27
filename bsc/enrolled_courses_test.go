package bsc

import (
	"testing"
	"time"
)

func TestParseTimeOfDay(t *testing.T) {
	strsAndValues := map[string]TimeOfDay{
		"9:30AM":  9*60 + 30,
		"9:30PM":  21*60 + 30,
		"10:26AM": 10*60 + 26,
		"12:13PM": 12*60 + 13,
		"12:13AM": 13,
	}

	for str, val := range strsAndValues {
		if res, err := ParseTimeOfDay(str); err != nil {
			t.Error("error for time "+str+":", err)
		} else if res != val {
			t.Error("bad result for: " + str)
		}
	}

	for _, errStr := range []string{"903:0AM", "1012:30AM", "10:30", "10:5AM"} {
		if _, err := ParseTimeOfDay(errStr); err == nil {
			t.Error("expected error for: " + errStr)
		}
	}
}

func TestParseWeeklyTimes(t *testing.T) {
	testWeeklyTimes(t, "MoWeFr 10:00AM - 10:50AM", WeeklyTimes{
		[]time.Weekday{time.Monday, time.Wednesday, time.Friday},
		TimeOfDay(10 * 60), TimeOfDay(10*60 + 50),
	})
	testWeeklyTimes(t, "Mo 12:30PM - 1:20PM", WeeklyTimes{
		[]time.Weekday{time.Monday},
		TimeOfDay(12*60 + 30), TimeOfDay(13*60 + 20),
	})
	badStrings := []string{
		"MoWeFri 10:10AM - 10:30AM",
		"MoWeFr 10:10AM 10:30AM",
		"MoWeFi 10:10AM - 10:30AM",
	}
	for _, errStr := range badStrings {
		if _, err := ParseWeeklyTimes(errStr); err == nil {
			t.Error("expected error for: " + errStr)
		}
	}
}

func TestTimeOfDayString(t *testing.T) {
	times := []string{"2:30AM", "2:05AM", "12:30AM", "12:30PM", "1:30PM"}
	for _, timeStr := range times {
		parsed, err := ParseTimeOfDay(timeStr)
		if err != nil {
			t.Error(err)
			continue
		}
		if parsed.String() != timeStr {
			t.Error("incorrect string for: " + timeStr)
		}
	}
}

func testWeeklyTimes(t *testing.T, str string, expect WeeklyTimes) {
	parsed, err := ParseWeeklyTimes(str)
	if err != nil {
		t.Error(err)
		return
	}
	if len(parsed.Days) != len(expect.Days) {
		t.Error("day count does not match for: " + str)
	} else {
		for i, d := range parsed.Days {
			if d != expect.Days[i] {
				t.Error("day at index", i, "does not match for: "+str)
			}
		}
	}
	if parsed.Start != expect.Start {
		t.Error("start does not match for: " + str)
	}
	if parsed.End != expect.End {
		t.Error("end does not match for: " + str)
	}
}
