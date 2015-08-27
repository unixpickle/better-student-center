package bsc

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// A Course represents a single course in which the user is enrolled.
// A course may contain multiple sections. For example, it could have a Lecture and a Discussion.
type Course struct {
	Name       string
	Department string
	Number     string
	Status     EnrollmentStatus
	Units      float64
	Components []Component
}

// A Component is one component of a course. Components have meeting times, locations, and
// instructors.
type Component struct {
	ClassNumber int
	Section     string
	Type        ComponentType
	WeeklyTimes WeeklyTimes
	Instructors []string
	Room        string
	StartDate   Date
	EndDate     Date
}

// A ComponentType represents the type of a Component. This may be, for example,
// ComponentTypeLecture or ComponentTypeDiscussion.
type ComponentType int

const (
	ComponentTypeLecture ComponentType = iota
	ComponentTypeDiscussion
	ComponentTypeLab
	ComponentTypeSeminar
	ComponentTypeRecitation
	ComponentTypeOther
)

// ParseComponentType turns a human-readable component type into a ComponentType. Unrecognized
// strings are treated as ComponentTypeOther.
func ParseComponentType(str string) ComponentType {
	// TODO: see if there are other component types.
	mapping := map[string]ComponentType{
		"Lecture":    ComponentTypeLecture,
		"Discussion": ComponentTypeDiscussion,
		"Laboratory": ComponentTypeLab,
		"Seminar":    ComponentTypeSeminar,
		"Recitation": ComponentTypeRecitation,
	}
	if ct, ok := mapping[str]; ok {
		return ct
	} else {
		return ComponentTypeOther
	}
}

// String generates a human-readable version of the ComponentType.
func (c ComponentType) String() string {
	names := map[ComponentType]string{
		ComponentTypeLecture:    "Lecture",
		ComponentTypeDiscussion: "Discussion",
		ComponentTypeLab:        "Laboratory",
		ComponentTypeSeminar:    "Seminar",
		ComponentTypeRecitation: "Recitation",
	}
	if name, ok := names[c]; ok {
		return name
	} else {
		return "Other"
	}
}

// EnrollmentStatus represents an enrollment status (e.g. dropped or enrolled) for a given course.
type EnrollmentStatus int

const (
	EnrollmentStatusEnrolled EnrollmentStatus = iota
	EnrollmentStatusDropped
	EnrollmentStatusWaitlisted
	EnrollmentStatusOther
)

// ParseEnrollmentStatus takes a human-readable string (e.g. "Enrolled") and turns it into an
// Enrollment status. Unrecognized strings are treated as EnrollmentStatusOther.
func ParseEnrollmentStatus(str string) EnrollmentStatus {
	// TODO: see if there are other enrollment statuses, and if "Waitlisted" is the correct string.
	mapping := map[string]EnrollmentStatus{
		"Enrolled":   EnrollmentStatusEnrolled,
		"Dropped":    EnrollmentStatusDropped,
		"Waitlisted": EnrollmentStatusWaitlisted,
	}
	if es, ok := mapping[str]; ok {
		return es
	} else {
		return EnrollmentStatusOther
	}
}

// String returns a human-readable (i.e. English-speaker readable) version of the enrollment status.
func (e EnrollmentStatus) String() string {
	names := map[EnrollmentStatus]string{
		EnrollmentStatusEnrolled:   "Enrolled",
		EnrollmentStatusDropped:    "Dropped",
		EnrollmentStatusWaitlisted: "Waitlisted",
	}
	if name, ok := names[e]; ok {
		return name
	} else {
		return "Other"
	}
}

// WeeklyTimes represents the weekly meeting times of a given section.
type WeeklyTimes struct {
	Days  []time.Weekday
	Start TimeOfDay
	End   TimeOfDay
}

// ParseWeeklyTimes parses a string like "MoWeFr 11:30AM - 12:30PM"
func ParseWeeklyTimes(times string) (*WeeklyTimes, error) {
	comps := strings.Split(times, " ")
	if len(comps) != 4 {
		return nil, errors.New("invalid weekly times: " + times)
	}
	if comps[2] != "-" {
		return nil, errors.New("missing separating dash: " + times)
	}

	start, err := ParseTimeOfDay(comps[1])
	if err != nil {
		return nil, err
	}
	end, err := ParseTimeOfDay(comps[3])
	if err != nil {
		return nil, err
	}
	days, err := parseWeekdays(comps[0])
	if err != nil {
		return nil, err
	}

	return &WeeklyTimes{days, start, end}, nil
}

// parseWeekdays parses a string like "MoWeFr" and turns it into an ordered list of weekdays.
func parseWeekdays(weekdays string) ([]time.Weekday, error) {
	if len(weekdays)%2 != 0 {
		return nil, errors.New("weekdays have invalid length: " + weekdays)
	}
	nameToWeekday := map[string]time.Weekday{
		"Mo": time.Monday,
		"Tu": time.Tuesday,
		"We": time.Wednesday,
		"Th": time.Thursday,
		"Fr": time.Friday,
	}
	res := make([]time.Weekday, len(weekdays)/2)
	for i := 0; i < len(weekdays); i += 2 {
		str := weekdays[i : i+2]
		if weekday, ok := nameToWeekday[str]; ok {
			res[i/2] = weekday
		} else {
			return nil, errors.New("invalid weekdays: " + weekdays)
		}
	}
	return res, nil
}

// A TimeOfDay represents a time of day as a number of minutes since 0:00.
type TimeOfDay int

// ParseTimeOfDay parses a 12-hour time.
// For example, this would turn 11:30AM into TimeOfDay(11*60 + 30) = TimeOfDay(690).
func ParseTimeOfDay(s string) (TimeOfDay, error) {
	if len(s) < 6 {
		return 0, errors.New("time is too short: " + s)
	} else if len(s) > 7 {
		return 0, errors.New("time is too long: " + s)
	}

	meridiemOffset := 0
	lastTwoLetters := s[len(s)-2:]
	if lastTwoLetters == "PM" {
		meridiemOffset = 12 * 60
	} else if lastTwoLetters != "AM" {
		return 0, errors.New("time must end with AM or PM: " + s)
	}

	if s[len(s)-5] != ':' {
		return 0, errors.New("time missing a ':' at the correct spot: " + s)
	}

	minuteStr := s[len(s)-4 : len(s)-2]
	hourStr := s[:len(s)-5]

	minuteNum, err := strconv.Atoi(minuteStr)
	if err != nil {
		return 0, err
	}

	hourNum, err := strconv.Atoi(hourStr)
	if err != nil {
		return 0, err
	}
	if hourNum == 12 {
		hourNum = 0
	}

	return TimeOfDay(minuteNum + hourNum*60 + meridiemOffset), nil
}

// Hour returns the hour component of a TimeOfDay in 24-hour time.
func (t TimeOfDay) Hour() int {
	return int(t) / 60
}

// Minute returns the minute component of a TimeOfDay.
func (t TimeOfDay) Minute() int {
	return int(t) % 60
}

// String returns a human-readable, 12-hour version of this time.
func (t TimeOfDay) String() string {
	hour := t.Hour()
	minute := t.Minute()
	amPmStr := "AM"
	if hour >= 12 {
		amPmStr = "PM"
	}

	if hour == 0 {
		hour = 12
	} else if hour > 12 {
		hour -= 12
	}

	minuteStr := strconv.Itoa(minute)
	if len(minuteStr) == 1 {
		minuteStr = "0" + minuteStr
	}

	return strconv.Itoa(hour) + ":" + minuteStr + amPmStr
}

// Date represents a day, given by a month, a day, and a year.
// The day starts at 1 to reflect the standard way of writing calendar dates.
type Date struct {
	Month time.Month
	Day   int
	Year  int
}

// ParseDate parses a slash-separated date string such as "05/08/2015".
func ParseDate(dateStr string) (date Date, err error) {
	comps := strings.Split(dateStr, "/")
	if len(comps) != 3 {
		return date, errors.New("string does not contain exactly two slashes: " + dateStr)
	}
	monthNum, err := strconv.Atoi(comps[0])
	if err != nil {
		return
	}
	date.Month = time.Month(monthNum)

	date.Day, err = strconv.Atoi(comps[1])
	if err != nil {
		return
	}

	date.Year, err = strconv.Atoi(comps[2])
	if err != nil {
		return
	}
	return
}

// String generates a slash-separated date string. It uses two digits for month and day, prepending
// zeroes as necessary.
func (d Date) String() string {
	monthStr := strconv.Itoa(int(d.Month))
	dayStr := strconv.Itoa(d.Day)
	yearStr := strconv.Itoa(d.Year)
	if len(dayStr) == 1 {
		dayStr = "0" + dayStr
	}
	if len(monthStr) == 1 {
		monthStr = "0" + monthStr
	}
	return monthStr + "/" + dayStr + "/" + yearStr
}
