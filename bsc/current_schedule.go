package bsc

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

// ParseCurrentSchedule parses the courses from the schedule list view page.
func ParseCurrentSchedule(page io.Reader) ([]Course, error) {
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

// parseComponentInfoMap processes a row from a courses's components and returns a Component with
// all the available information.
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
	component.Type = ParseComponentType(infoMap["Component"])

	component.Instructors = strings.Split(infoMap["Instructor"], ",")
	for i, instructor := range component.Instructors {
		component.Instructors[i] = strings.TrimSpace(instructor)
	}

	return
}

// parseCourseInfoTable takes a table with general course fields and turns it into a Course. This
// will not fill in certain fields (i.e. the components and name of the course).
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
	course.Status = ParseEnrollmentStatus(infoMap["Status"])

	return
}
