package bsc

import "time"

var enrolledCoursesPath string = "EMPLOYEE/HRMS/c/SA_LEARNER_SERVICES.SSR_SSENRL_LIST.GBL" +
	"?Page=SSR_SSENRL_LIST"

// A ComponentType represents the type of a section. This may be, for example, ComponentTypeLecture
// or ComponentTypeDiscussion.
type ComponentType int

const (
	ComponentTypeLecture    = 0
	ComponentTypeDiscussion = 1
	ComponentTypeOther      = 2
)

// A TimeOfDay represents a number of minutes since 0:00.
type TimeOfDay int

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
	Number     int
	Enrolled   bool
	Units      float64
	Graded     bool
	Sections   []Section
}

// Times represents the weekly meeting times of a given section.
type WeeklyTimes struct {
	Days []time.Weekday
	Time TimeOfDay
}

// A Section is one component of a course. Sections have meeting times, locations, and instructors.
type Section struct {
	ClassNumber int
	Number      int
	Type        ComponentType
	WeeklyTimes WeeklyTimes
	Instructors []string
	Room        string
	StartDate   time.Time
	EndDate     time.Time
}
