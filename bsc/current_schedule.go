package bsc

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// fetchExtraScheduleInfo gets more information about each component.
//
// The rootNode argument should be the parsed schedule list view.
func fetchExtraScheduleInfo(client *http.Client, courses []Course, rootNode *html.Node) error {
	psForm, ok := scrape.Find(rootNode, scrape.ByClass("PSForm"))
	if !ok {
		return errors.New("could not find PSForm")
	}
	icsid, ok := scrape.Find(psForm, scrape.ById("ICSID"))
	if !ok {
		return errors.New("could not find ICSID")
	}

	formAction := getNodeAttribute(psForm, "action")
	sid := getNodeAttribute(icsid, "value")

	wg := sync.WaitGroup{}

	var err error
	var errLock sync.Mutex

	sectionIndex := 0
	for _, course := range courses {
		for componentIndex := range course.Components {
			component := &course.Components[componentIndex]
			wg.Add(1)
			go func(setCourseOpen bool, index int, comp *Component) {
				defer wg.Done()

				postData := generateClassDetailForm(sid, index)
				res, reqError := client.PostForm(formAction, postData)
				if res != nil {
					defer res.Body.Close()
				}
				if reqError != nil {
					errLock.Lock()
					err = reqError
					errLock.Unlock()
					return
				}

				courseOpen, parseErr := parseExtraComponentInfo(res.Body, component)
				if parseErr != nil {
					fmt.Println("guy failed,", parseErr)
					errLock.Lock()
					err = parseErr
					errLock.Unlock()
					return
				}

				if setCourseOpen {
					course.Open = &courseOpen
				}
			}(componentIndex == 0, sectionIndex, component)
			sectionIndex++
		}
	}

	wg.Wait()
	return err
}

// generateClassDetailForm generates the POST values for extended component information.
func generateClassDetailForm(icsid string, sectionIndex int) url.Values {
	postData := url.Values{}
	for _, f := range []string{"ICFocus", "ICFind", "ICAddCount", "ICAPPCLSDATA"} {
		postData.Add(f, "")
	}

	// TODO: compress this code using loops and such.

	postData.Add("ICAJAX", "1")
	postData.Add("ICNAVTYPEDROPDOWN", "0")
	postData.Add("ICType", "Panel")
	postData.Add("ICElementNum", "0")
	postData.Add("ICStateNum", "4")
	postData.Add("ICAction", "MTG_SECTION$"+strconv.Itoa(sectionIndex))
	postData.Add("ICXPos", "0")
	postData.Add("ICYPos", "179")
	postData.Add("ResponsetoDiffFrame", "-1")
	postData.Add("TargetFrameName", "None")
	postData.Add("FacetPath", "None")
	postData.Add("ICSaveWarningFilter", "0")
	postData.Add("ICChanged", "-1")
	postData.Add("ICResubmit", "0")
	postData.Add("ICSID", icsid)
	postData.Add("ICActionPrompt", "false")
	postData.Add("DERIVED_SSTSNAV_SSTS_MAIN_GOTO$7$", "9999")
	postData.Add("DERIVED_REGFRM1_SSR_SCHED_FORMAT$258$", "L")
	postData.Add("DERIVED_REGFRM1_SA_STUDYLIST_E$chk", "Y")
	postData.Add("DERIVED_REGFRM1_SA_STUDYLIST_E", "Y")
	postData.Add("DERIVED_REGFRM1_SA_STUDYLIST_D$chk", "Y")
	postData.Add("DERIVED_REGFRM1_SA_STUDYLIST_D", "Y")
	postData.Add("DERIVED_REGFRM1_SA_STUDYLIST_W$chk", "Y")
	postData.Add("DERIVED_REGFRM1_SA_STUDYLIST_W", "Y")
	postData.Add("DERIVED_SSTSNAV_SSTS_MAIN_GOTO$8$", "9999")

	return postData
}

// parseCurrentSchedule parses the courses from the schedule list view page.
//
// If fetchMoreInfo is true, this will perform a request for each component to find out information
// about it.
func parseCurrentSchedule(rootNode *html.Node) ([]Course, error) {
	courseTables := scrape.FindAll(rootNode, scrape.ByClass("PSGROUPBOXWBO"))
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

	startEndComps := strings.Split(infoMap["Start/End Date"], " - ")
	if len(startEndComps) != 2 {
		err = errors.New("invalid start/end date: " + infoMap["Start/End Date"])
		return
	}
	if component.StartDate, err = ParseDate(startEndComps[0]); err != nil {
		return
	}
	if component.EndDate, err = ParseDate(startEndComps[1]); err != nil {
		return
	}

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

// parseExtraComponentInfo parses the "Class Detail" page for a component.
func parseExtraComponentInfo(body io.Reader, component *Component) (courseOpen bool, err error) {
	nodes, err := html.ParseFragment(body, nil)
	if err != nil {
		return
	}
	if len(nodes) != 1 {
		return false, errors.New("invalid number of root elements")
	}

	openStatus, ok := scrape.Find(nodes[0], scrape.ById("SSR_CLS_DTL_WRK_SSR_DESCRSHORT"))
	if !ok {
		fmt.Println("failed for", component)
		return false, errors.New("open status not found")
	}
	courseOpen = (nodeInnerText(openStatus) == "Open")

	availTable, ok := scrape.Find(nodes[0], scrape.ById("ACE_SSR_CLS_DTL_WRK_GROUP3"))
	if !ok {
		return courseOpen, errors.New("could not find availability info")
	}

	rows := scrape.FindAll(availTable, scrape.ByTag(atom.Tr))
	if len(rows) != 7 {
		return courseOpen, errors.New("invalid number of rows in availability table")
	}

	var availability ClassAvailability

	cols := nodesWithAlignAttribute(scrape.FindAll(rows[2], scrape.ByTag(atom.Td)))
	if len(cols) != 2 {
		return courseOpen, errors.New("expected 2 aligned columns in row 2")
	}
	availability.Capacity, err = strconv.Atoi(strings.TrimSpace(nodeInnerText(cols[0])))
	if err != nil {
		return
	}
	availability.WaitListCapacity, err = strconv.Atoi(strings.TrimSpace(nodeInnerText(cols[1])))
	if err != nil {
		return
	}

	cols = nodesWithAlignAttribute(scrape.FindAll(rows[4], scrape.ByTag(atom.Td)))
	if len(cols) != 2 {
		return courseOpen, errors.New("expected 2 aligned columns in row 4")
	}
	availability.EnrollmentTotal, err = strconv.Atoi(strings.TrimSpace(nodeInnerText(cols[0])))
	if err != nil {
		return
	}
	availability.WaitListTotal, err = strconv.Atoi(strings.TrimSpace(nodeInnerText(cols[1])))
	if err != nil {
		return
	}

	cols = nodesWithAlignAttribute(scrape.FindAll(rows[6], scrape.ByTag(atom.Td)))
	if len(cols) != 1 {
		return courseOpen, errors.New("expected 1 aligned column in row 6")
	}
	availability.AvailableSeats, err = strconv.Atoi(strings.TrimSpace(nodeInnerText(cols[0])))
	if err != nil {
		return
	}

	component.ClassAvailability = &availability

	return
}
