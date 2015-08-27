package bsc

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

var enrolledCoursesPath string = "/EMPLOYEE/HRMS/c/SA_LEARNER_SERVICES.SSR_SSENRL_LIST.GBL" +
	"?Page=SSR_SSENRL_LIST"

// A ComponentType represents the type of a Component. This may be, for example,
// ComponentTypeLecture or ComponentTypeDiscussion.
type ComponentType int

const (
	ComponentTypeLecture ComponentType = iota
	ComponentTypeDiscussion
	ComponentTypeOther
)

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

// A Course represents a single course in which the user is enrolled.
// A course may contain multiple sections. For example, it could have a Lecture and a Discussion.
type Course struct {
	Name       string
	Department string
	Number     string
	Enrolled   bool
	Units      float64
	Graded     bool
	Components []Component
}

// ParseCourses parses the class schedule list view.
func ParseCourses(page io.Reader) ([]Course, error) {
	nodes, err := html.ParseFragment(page, nil)
	if err != nil {
		return nil, err
	}
	if len(nodes) != 1 {
		return nil, errors.New("invalid number of root elements")
	}

	courseTables := scrape.FindAll(nodes[0], scrape.ByClass("PSGROUPBOXWBO"))
	result := make([]Course, 0, len(courseTables))
	for _, classTable := range courseTables {
		var course Course
		titleElement, ok := scrape.Find(classTable, scrape.ByClass("PAGROUPDIVIDER"))
		if !ok {
			// This will occur at least once, since the filter options are a PSGROUPBOXWBO.
			continue
		}

		// NOTE: there isn't really a standard way to parse the department/number.
		course.Name = nodeInnerText(titleElement)

		infoTables := scrape.FindAll(classTable, scrape.ByClass("PSLEVEL3GRIDNBO"))
		if len(infoTables) != 2 {
			return nil, errors.New("expected exactly 2 info tables but found " +
				strconv.Itoa(len(infoTables)))
		}

		courseInfoTable := infoTables[0]
		courseInfoEntries, err := tableEntriesAsMaps(courseInfoTable)
		if err != nil {
			return nil, err
		}
		if len(courseInfoEntries) != 1 {
			return nil, errors.New("expected exactly 1 row in the course info table but got " +
				strconv.Itoa(len(courseInfoEntries)))
		}
		courseInfoMap := courseInfoEntries[0]

		if unitsStr, ok := courseInfoMap["Units"]; ok {
			course.Units, _ = strconv.ParseFloat(unitsStr, -1)
		}

		// TODO: figure out other Status values.
		if courseInfoMap["Status"] == "Enrolled" {
			course.Enrolled = true
		}

		// TODO: parse the components.
		// componentInfoTable := infoTables[1]

		result = append(result, course)
	}

	return result, nil
}

// Times represents the weekly meeting times of a given section.
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

// A Component is one component of a course. Components have meeting times, locations, and
// instructors.
type Component struct {
	ClassNumber int
	Number      int
	Type        ComponentType
	WeeklyTimes WeeklyTimes
	Instructors []string
	Room        string
	StartDate   time.Time
	EndDate     time.Time
}

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
