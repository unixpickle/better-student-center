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
	ComponentTypeLab
	ComponentTypeOther
)

// EnrollmentStatus represents an enrollment status (e.g. dropped or enrolled) for a given course.
type EnrollmentStatus int

// String returns a human-readable (i.e. English-speaker readable) version of the enrollment status.
func (e EnrollmentStatus) String() string {
	names := map[EnrollmentStatus]string{
		EnrollmentStatusEnrolled: "Enrolled",
		EnrollmentStatusDropped: "Dropped",
		EnrollmentStatusWaitlisted: "Waitlisted",
	}
	if name, ok := names[e]; ok {
		return name
	} else {
		return "Other"
	}
}

const (
	EnrollmentStatusEnrolled EnrollmentStatus = iota
	EnrollmentStatusDropped
	EnrollmentStatusWaitlisted
	EnrollmentStatusOther
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
		titleElement, ok := scrape.Find(classTable, scrape.ByClass("PAGROUPDIVIDER"))
		if !ok {
			// This will occur at least once, since the filter options are a PSGROUPBOXWBO.
			continue
		}

		infoTables := scrape.FindAll(classTable, scrape.ByClass("PSLEVEL3GRIDNBO"))
		if len(infoTables) != 2 {
			return nil, errors.New("expected exactly 2 info tables but found " +
				strconv.Itoa(len(infoTables)))
		}

		courseInfoTable := infoTables[0]
		course, err := parseCourseInfoTable(courseInfoTable)
		if err != nil {
			return nil, err
		}

		// NOTE: there isn't really a standard way to parse the department/number.
		course.Name = nodeInnerText(titleElement)

		componentsInfoTable := infoTables[1]
		componentMaps, err := tableEntriesAsMaps(componentsInfoTable)
		if err != nil {
			return nil, err
		}
		course.Components = make([]Component, len(componentMaps))
		for i, componentMap := range componentMaps {
			course.Components[i], err = parseComponentInfoMap(componentMap)
			if err != nil {
				return nil, err
			}
		}

		result = append(result, course)
	}

	return result, nil
}

func parseComponentInfoMap(infoMap map[string]string) (component Component, err error) {
	component.ClassNumber, err = strconv.Atoi(infoMap["Class Nbr"])
	if err != nil {
		return
	}

	weeklyTimes, err := ParseWeeklyTimes(infoMap["Days & Times"])
	if err != nil {
		return
	} else {
		component.WeeklyTimes = *weeklyTimes
	}

	// TODO: parse start/end date.

	component.Section = infoMap["Section"]
	component.Room = infoMap["Room"]

	component.Instructors = strings.Split(infoMap["Instructor"], ",")
	for i, instructor := range component.Instructors {
		component.Instructors[i] = strings.TrimSpace(instructor)
	}

	// TODO: see if there are more possible component types.
	switch infoMap["Component"] {
	case "Lecture":
		component.Type = ComponentTypeLecture
	case "Discussion":
		component.Type = ComponentTypeDiscussion
	case "Lab":
		// TODO: see if this is a real component type.
		component.Type = ComponentTypeLab
	default:
		component.Type = ComponentTypeOther
	}

	return
}

func parseCourseInfoTable(table *html.Node) (course Course, err error) {
	infoMaps, err := tableEntriesAsMaps(table)
	if err != nil {
		return
	}
	if len(infoMaps) != 1 {
		return course, errors.New("expected exactly 1 row in the course info table but got " +
			strconv.Itoa(len(infoMaps)))
	}
	infoMap := infoMaps[0]

	// TODO: figure out how to use the "Graded" field in a universal way. The string for this may
	// differ between universities.

	if unitsStr, ok := infoMap["Units"]; ok {
		course.Units, _ = strconv.ParseFloat(unitsStr, -1)
	}

	// TODO: verify these Status strings and make sure there are no others.
	switch infoMap["Status"] {
	case "Enrolled":
		course.Status = EnrollmentStatusEnrolled
	case "Dropped":
		course.Status = EnrollmentStatusDropped
	case "Waitlisted":
		course.Status = EnrollmentStatusWaitlisted
	default:
		course.Status = EnrollmentStatusOther
	}

	return
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
	Section     string
	Type        ComponentType
	WeeklyTimes WeeklyTimes
	Instructors []string
	Room        string
	StartDate   time.Time
	EndDate     time.Time
}

func parseWeekdays(weekdays string) ([]time.Weekday, error) {
	// TODO: move this func to a more appropriate place in the file.
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
